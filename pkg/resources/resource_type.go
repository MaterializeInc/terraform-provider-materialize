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

var typeSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("type", true, true),
	"schema_name":        SchemaNameSchema("type", false),
	"database_name":      DatabaseNameSchema("type", false),
	"qualified_sql_name": QualifiedNameSchema("type"),
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
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"map_properties"},
		AtLeastOneOf:  []string{"map_properties", "list_properties"},
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
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"list_properties"},
		AtLeastOneOf:  []string{"map_properties", "list_properties"},
	},
	"category": {
		Description: "Type category.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
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

		Schema: typeSchema,
	}
}

func typeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	s, err := materialize.ScanType(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

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

	return nil
}

func typeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	typeName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.MaterializeObject{ObjectType: "TYPE", Name: typeName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewTypeBuilder(meta.(*sqlx.DB), o)

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

	// set id
	i, err := materialize.TypeId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

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
