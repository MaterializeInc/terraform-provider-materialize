package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var typeSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("type", true, true),
	"schema_name":        SchemaNameSchema("type", false),
	"database_name":      DatabaseNameSchema("type", false),
	"qualified_sql_name": QualifiedNameSchema("type"),
	"comment":            CommentSchema(false),
	"row_properties": {
		Description: "Row properties.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_name": {
					Description: "The name of a field in a row type.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"field_type": {
					Description: "The data type of a field indicated by `FIELD NAME`.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		Optional:     true,
		MinItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"row_properties", "map_properties", "list_properties"},
	},
	"list_properties": {
		Description: "List properties.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"element_type": {
					Description: "Creates a custom list whose elements are of `ELEMENT TYPE`",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		Optional:     true,
		MinItems:     1,
		MaxItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"row_properties", "map_properties", "list_properties"},
	},
	"map_properties": {
		Description: "Map properties.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key_type": {
					Description: "Creates a custom map whose keys are of `KEY TYPE`. `KEY TYPE` must resolve to text.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"value_type": {
					Description: "Creates a custom map whose values are of `VALUE TYPE`.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		Optional:     true,
		MinItems:     1,
		MaxItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"row_properties", "map_properties", "list_properties"},
	},
	"category": {
		Description: "Type category.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
}

// Define the V0 schema function
func typeSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: typeSchema,
	}
}

func Type() *schema.Resource {
	return &schema.Resource{
		Description: "A custom types, which let you create named versions of anonymous types.",

		CreateContext: typeCreate,
		ReadContext:   typeRead,
		UpdateContext: typeUpdate,
		DeleteContext: typeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        typeSchema,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    typeSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: utils.IdStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func typeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	s, err := materialize.ScanType(meta.(*sqlx.DB), utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(i))

	if err := d.Set("name", s.TypeName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("category", s.Category.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.TypeName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func typeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	typeName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "TYPE", Name: typeName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewTypeBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("row_properties"); ok {
		p := materialize.GetRowProperties(v)
		b.RowProperties(p)
	}

	if v, ok := d.GetOk("list_properties"); ok {
		p := materialize.GetListProperties(v)
		b.ListProperties(p)
	}

	if v, ok := d.GetOk("map_properties"); ok {
		p := materialize.GetMapProperties(v)
		b.MapProperties(p)
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

	// set id
	i, err := materialize.TypeId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(i))

	return typeRead(ctx, d, meta)
}

func typeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	typeName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "TYPE", Name: typeName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")

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

	return typeRead(ctx, d, meta)
}

func typeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	typeName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{Name: typeName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewTypeBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
