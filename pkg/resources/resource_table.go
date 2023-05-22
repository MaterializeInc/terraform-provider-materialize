package resources

import (
	"context"
	"log"

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
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadTableParams(i)

	var name, schema, database *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
		if err == sql.ErrNoRows {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
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

	b := materialize.NewTableBuilder(*name, *schema, *database)
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func tableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewTableBuilder(tableName, schemaName, databaseName)

	if v, ok := d.GetOk("column"); ok {
		columns := materialize.GetTableColumnStruct(v.([]interface{}))
		builder.Column(columns)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "table"); err != nil {
		return diag.FromErr(err)
	}
	return tableRead(ctx, d, meta)
}

func tableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := materialize.NewTableBuilder(tableName, schemaName, databaseName).Rename(newName.(string))

		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename table: %s", q)
			return diag.FromErr(err)
		}
	}

	return tableRead(ctx, d, meta)
}

func tableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewTableBuilder(tableName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "table"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
