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

var viewSchema = map[string]*schema.Schema{
	"name":               NameSchema("view", true, false),
	"schema_name":        SchemaNameSchema("view", false),
	"database_name":      DatabaseNameSchema("view", false),
	"qualified_sql_name": QualifiedNameSchema("view"),
	"statement": {
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

type ViewParams struct {
	ViewName     sql.NullString `db:"view_name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
}

func viewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadViewParams(i)

	var s ViewParams
	if err := conn.Get(&s, q); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ViewName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewViewBuilder(s.ViewName.String, s.SchemaName.String, s.DatabaseName.String)
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func viewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewViewBuilder(viewName, schemaName, databaseName)

	if v, ok := d.GetOk("statement"); ok && v.(string) != "" {
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

		q := materialize.NewViewBuilder(viewName, schemaName, databaseName).Rename(newName.(string))

		if err := execResource(conn, q); err != nil {
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

	q := materialize.NewViewBuilder(viewName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "view"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
