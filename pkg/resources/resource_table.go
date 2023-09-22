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

var tableSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("table", true, false),
	"schema_name":        SchemaNameSchema("table", false),
	"database_name":      DatabaseNameSchema("table", false),
	"qualified_sql_name": QualifiedNameSchema("table"),
	"comment":            CommentSchema(),
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
					StateFunc: func(val any) string {
						alias, ok := aliases[val.(string)]
						if ok {
							return alias
						}
						return val.(string)
					},
				},
				"nullable": {
					Description: "Do not allow the column to contain NULL values. Columns without this constraint can contain NULL values.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"comment": CommentSchema(),
			},
		},
		Optional: true,
		MinItems: 1,
		ForceNew: true,
	},
	"ownership_role": OwnershipRoleSchema(),
}

func Table() *schema.Resource {
	return &schema.Resource{
		Description: "A table persists durable storage that can be written to, updated and seamlessly joined with other tables, views or sources",

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

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.TableName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	// Table columns
	tableColumns, err := materialize.ListTableColumns(meta.(*sqlx.DB), i)
	if err != nil {
		log.Print("[DEBUG] cannot query list tables")
		return diag.FromErr(err)
	}
	var tc []interface{}
	for _, t := range tableColumns {
		column := map[string]interface{}{"name": t.Name.String, "type": t.Type.String, "nullable": !t.Nullable.Bool, "comment": t.Comment.String}
		tc = append(tc, column)
	}
	if err := d.Set("column", tc); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func tableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewTableBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("column"); ok {
		columns := materialize.GetTableColumnStruct(v.([]interface{}))
		b.Column(columns)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// column comment
	if v, ok := d.GetOk("column"); ok {
		columns := materialize.GetTableColumnStruct(v.([]interface{}))
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		for _, c := range columns {
			if c.Comment != "" {
				if err := comment.Column(c.ColName, c.Comment); err != nil {
					log.Printf("[DEBUG] resource failed column comment, dropping object: %s", o.Name)
					b.Drop()
					return diag.FromErr(err)
				}
			}
		}
	}

	// set id
	i, err := materialize.TableId(meta.(*sqlx.DB), o)
	if err != nil {
		log.Printf("[DEBUG] cannot query table: %s", o.QualifiedName())
		return diag.FromErr(err)
	}

	d.SetId(i)

	return tableRead(ctx, d, meta)
}

func tableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "TABLE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewTableBuilder(meta.(*sqlx.DB), o)

		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("columns") {
		_, newColumns := d.GetChange("columns")
		columns := materialize.GetTableColumnStruct(newColumns.([]interface{}))
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		// Reset all comments if change present
		for _, c := range columns {
			if c.Comment != "" {
				if err := comment.Column(c.ColName, c.Comment); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return tableRead(ctx, d, meta)
}

func tableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewTableBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
