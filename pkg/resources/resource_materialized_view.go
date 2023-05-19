package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var materializedViewSchema = map[string]*schema.Schema{
	"name":               NameSchema("materialized view", true, false),
	"schema_name":        SchemaNameSchema("materialized view", false),
	"database_name":      DatabaseNameSchema("materialized view", false),
	"qualified_sql_name": QualifiedNameSchema("materialized view"),
	"cluster_name": {
		Description: "The cluster to maintain the materialized view. If not specified, defaults to the default cluster.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"statement": {
		Description: "The SQL statement to create the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
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

type MaterializedViewParams struct {
	MaterializedViewName sql.NullString `db:"name"`
	SchemaName           sql.NullString `db:"schema"`
	DatabaseName         sql.NullString `db:"database"`
	Cluster              sql.NullString `db:"cluster"`
}

func materializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadMaterializedViewParams(i)

	var s MaterializedViewParams
	if err := conn.Get(&s, q); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.MaterializedViewName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.Cluster.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.MaterializedViewName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func materializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok && v.(string) != "" {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("statement"); ok && v.(string) != "" {
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

		q := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName).Rename(newName.(string))

		if err := execResource(conn, q); err != nil {
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

	q := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "materialized view"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
