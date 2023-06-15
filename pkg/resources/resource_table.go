package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var tableSchema = map[string]*schema.Schema{
	"name":               NameSchema("table", true, false),
	"schema_name":        SchemaNameSchema("table", false),
	"database_name":      DatabaseNameSchema("table", false),
	"qualified_sql_name": QualifiedNameSchema("table"),
	"column": {
		Description: "Column of the table.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the column to be created in the table.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"type": {
					Description: "The data type of the column indicated by name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"nullable": {
					Description: "	Do not allow the column to contain NULL values. Columns without this constraint can contain NULL values.",
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
		Optional: true,
		MinItems: 1,
		ForceNew: true,
	},
}

func Table() *schema.Resource {
	return &schema.Resource{
		Description: "A table persists in durable storage and can be written to, updated and seamlessly joined with other tables, views or sources.",

		CreateContext: tableCreate,
		ReadContext:   tableRead,
		UpdateContext: tableUpdate,
		DeleteContext: tableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: tableSchema,
	}
}

func tableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	s, err := materialize.ScanTable(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.TableName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.TableName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func tableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewTableBuilder(meta.(*sqlx.DB), tableName, schemaName, databaseName)

	if v, ok := d.GetOk("column"); ok {
		columns := materialize.GetTableColumnStruct(v.([]interface{}))
		b.Column(columns)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.TableId(meta.(*sqlx.DB), tableName, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return tableRead(ctx, d, meta)
}

func tableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		b := materialize.NewTableBuilder(meta.(*sqlx.DB), oldName.(string), schemaName, databaseName)

		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return tableRead(ctx, d, meta)
}

func tableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewTableBuilder(meta.(*sqlx.DB), tableName, schemaName, databaseName)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
