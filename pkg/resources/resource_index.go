package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description:  "The identifier for the index.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"schema_name": {
		Description: "The identifier for the index schema.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"database_name": {
		Description: "The identifier for the index database.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"default": {
		Description:  "Creates a default index using all inferred columns are used.",
		Type:         schema.TypeBool,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"qualified_sql_name": QualifiedNameSchema("index"),
	"comment":            CommentSchema(false),
	"obj_name":           IdentifierSchema("obj_name", "The name of the source, view, or materialized view on which you want to create an index.", true),
	"cluster_name": {
		Description: "The cluster to maintain this index. If not specified, defaults to the active cluster.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"method": {
		Description:  "The name of the index method to use.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Default:      "ARRANGEMENT",
		ValidateFunc: validation.StringInSlice([]string{"ARRANGEMENT"}, true),
	},
	"col_expr": {
		Description: "The expressions to use as the key for the index.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Description: "The name of the option you want to set.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		Required: true,
		ForceNew: true,
	},
}

// Define the V0 schema function
func indexSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: indexSchema,
	}
}

func Index() *schema.Resource {
	return &schema.Resource{
		Description: "Indexes represent query results stored in memory.",

		CreateContext: indexCreate,
		ReadContext:   indexRead,
		UpdateContext: indexUpdate,
		DeleteContext: indexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        indexSchema,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    indexSchemaV0().CoreConfigSchema().ImpliedType(),
				Upgrade: utils.IdStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	s, err := materialize.ScanIndex(meta.(*sqlx.DB), utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(i))

	if err := d.Set("name", s.IndexName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.ObjectSchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.ObjectDatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.ObjectDatabaseName.String, s.ObjectSchemaName.String, s.IndexName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	// Index columns
	indexColumns, err := materialize.ListIndexColumns(meta.(*sqlx.DB), i)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(indexColumns) > 0 {
		var ic []interface{}
		for _, i := range indexColumns {
			column := map[string]interface{}{"field": i.Name.String}
			ic = append(ic, column)
		}
		if err := d.Set("col_expr", ic); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func indexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	indexName := d.Get("name").(string)
	indexDefault := d.Get("default").(bool)

	obj := d.Get("obj_name").([]interface{})[0].(map[string]interface{})

	o := materialize.MaterializeObject{ObjectType: "INDEX", Name: indexName}
	b := materialize.NewIndexBuilder(
		meta.(*sqlx.DB),
		o,
		indexDefault,
		materialize.IdentifierSchemaStruct{
			Name:         obj["name"].(string),
			SchemaName:   obj["schema_name"].(string),
			DatabaseName: obj["database_name"].(string),
		},
	)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("method"); ok {
		b.Method(v.(string))
	}

	if v, ok := d.GetOk("col_expr"); ok {
		c := materialize.GetIndexColumnStruct(v.([]interface{}))
		b.ColExpr(c)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		if err := b.Comment(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.IndexId(meta.(*sqlx.DB), indexName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(i))

	return indexRead(ctx, d, meta)
}

func indexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	indexName := d.Get("name").(string)
	o := materialize.MaterializeObject{ObjectType: "INDEX", Name: indexName}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return indexRead(ctx, d, meta)
}

func indexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	obj := d.Get("obj_name").([]interface{})[0].(map[string]interface{})
	name := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "INDEX", Name: name}
	b := materialize.NewIndexBuilder(
		meta.(*sqlx.DB),
		o,
		d.Get("default").(bool),
		materialize.IdentifierSchemaStruct{
			Name:         obj["name"].(string),
			SchemaName:   obj["schema_name"].(string),
			DatabaseName: obj["database_name"].(string),
		},
	)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
