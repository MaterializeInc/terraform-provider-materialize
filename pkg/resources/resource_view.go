package resources

import (
	"context"
	"database/sql"

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

func viewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	s, err := materialize.ScanView(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
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

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.ViewName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func viewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewViewBuilder(meta.(*sqlx.DB), viewName, schemaName, databaseName)

	if v, ok := d.GetOk("statement"); ok && v.(string) != "" {
		b.SelectStmt(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.ViewId(meta.(*sqlx.DB), viewName, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return viewRead(ctx, d, meta)
}

func viewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewViewBuilder(meta.(*sqlx.DB), viewName, schemaName, databaseName)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return viewRead(ctx, d, meta)
}

func viewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	viewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewViewBuilder(meta.(*sqlx.DB), viewName, schemaName, databaseName)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
