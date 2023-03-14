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

var viewSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the view.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the view schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
		ForceNew:    true,
	},
	"database_name": {
		Description: "The identifier for the view database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
		ForceNew:    true,
	},
	"qualified_name": {
		Description: "The fully qualified name of the view.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"select_stmt": {
		Description: "The SQL statement to create the view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func View() *schema.Resource {
	return &schema.Resource{
		Description: "A non-materialized view, provides an alias for the embedded SELECT statement.",

		CreateContext: viewCreate,
		ReadContext:   viewRead,
		UpdateContext: viewUpdate,
		DeleteContext: viewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: viewSchema,
	}
}

type ViewBuilder struct {
	viewName     string
	schemaName   string
	databaseName string
	selectStmt   string
}

func (b *ViewBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.viewName)
}

func newViewBuilder(viewName, schemaName, databaseName string) *ViewBuilder {
	return &ViewBuilder{
		viewName:     viewName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *ViewBuilder) SelectStmt(selectStmt string) *ViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *ViewBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	q.WriteString(fmt.Sprintf(` VIEW %s`, b.qualifiedName()))

	q.WriteString(` AS `)

	q.WriteString(b.selectStmt)

	q.WriteString(`;`)

	return q.String()
}

func (b *ViewBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER VIEW %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ViewBuilder) Drop() string {
	return fmt.Sprintf(`DROP VIEW %s;`, b.qualifiedName())
}

func (b *ViewBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_views.id
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_views.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.viewName, b.schemaName, b.databaseName)
}

func readViewParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_views.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_views.id = '%s';`, id)
}

func viewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readViewParams(i)

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

func viewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newViewBuilder(viewName, schemaName, databaseName)

	if v, ok := d.GetOk("select_stmt"); ok && v.(string) != "" {
		builder.SelectStmt(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "view"); err != nil {
		return diag.FromErr(err)
	}
	return viewRead(ctx, d, meta)
}

func viewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := newViewBuilder(viewName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename view: %s", q)
			return diag.FromErr(err)
		}
	}

	return viewRead(ctx, d, meta)
}

func viewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newViewBuilder(viewName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "view"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
