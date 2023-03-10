package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var materializedViewSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the materialized view schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
		ForceNew:    true,
	},
	"database_name": {
		Description: "The identifier for the materialized view database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
		ForceNew:    true,
	},
	"qualified_name": {
		Description: "The fully qualified name of the materialized view.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"in_cluster": {
		Description: "The cluster to maintain this materialized view.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"select_stmt": {
		Description: "The SQL statement to create the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
	},
}

func MaterializedView() *schema.Resource {
	return &schema.Resource{
		Description: "A materialized view persists in durable storage and can be written to, updated and seamlessly joined with other views, views or sources.",

		CreateContext: materializedViewCreate,
		ReadContext:   materializedViewRead,
		UpdateContext: materializedViewUpdate,
		DeleteContext: materializedViewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: materializedViewSchema,
	}
}

type MaterializedViewBuilder struct {
	materializedViewName string
	schemaName           string
	databaseName         string
	inCluster            string
	selectStmt           string
}

func (b *MaterializedViewBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.materializedViewName)
}

func newMaterializedViewBuilder(materializedViewName, schemaName, databaseName string) *MaterializedViewBuilder {
	return &MaterializedViewBuilder{
		materializedViewName: materializedViewName,
		schemaName:           schemaName,
		databaseName:         databaseName,
	}
}

func (b *MaterializedViewBuilder) InCluster(inCluster string) *MaterializedViewBuilder {
	b.inCluster = inCluster
	return b
}

func (b *MaterializedViewBuilder) SelectStmt(selectStmt string) *MaterializedViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *MaterializedViewBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	q.WriteString(fmt.Sprintf(` MATERIALIZED VIEW %s`, b.qualifiedName()))

	if b.inCluster != "" {
		q.WriteString(` IN CLUSTER `)
		q.WriteString(b.inCluster)
	}

	q.WriteString(` AS `)

	q.WriteString(b.selectStmt)

	q.WriteString(`;`)

	return q.String()
}

func (b *MaterializedViewBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER MATERIALIZED VIEW %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *MaterializedViewBuilder) Drop() string {
	return fmt.Sprintf(`DROP MATERIALIZED VIEW %s;`, b.qualifiedName())
}

func (b *MaterializedViewBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_materialized_views.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.materializedViewName, b.schemaName, b.databaseName)
}

func readMaterializedViewParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_materialized_views.id = '%s';`, id)
}

func materializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readMaterializedViewParams(i)

	var name, schema, database *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func materializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newMaterializedViewBuilder(materializedViewName, schemaName, databaseName)

	if v, ok := d.GetOk("in_cluster"); ok && v.(string) != "" {
		builder.InCluster(v.(string))
	}

	if v, ok := d.GetOk("select_stmt"); ok && v.(string) != "" {
		builder.SelectStmt(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "materialized view"); err != nil {
		return diag.FromErr(err)
	}
	return materializedViewRead(ctx, d, meta)
}

func materializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := newSecretBuilder(materializedViewName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename materialized view: %s", q)
			return diag.FromErr(err)
		}
	}

	return materializedViewRead(ctx, d, meta)
}

func materializedViewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newMaterializedViewBuilder(materializedViewName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "materialized view"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
