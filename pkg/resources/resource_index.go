package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

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
	"default": {
		Description:  "Creates a default index using all inferred columns are used.",
		Type:         schema.TypeBool,
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
	"qualified_sql_name": QualifiedNameSchema("view"),
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
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"default"},
	},
}

func Index() *schema.Resource {
	return &schema.Resource{
		Description: "An in-memory index on a source, view, or materialized view.",

		CreateContext: indexCreate,
		ReadContext:   indexRead,
		DeleteContext: indexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: indexSchema,
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	s, err := materialize.ScanIndex(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.IndexName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.Object.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.IndexName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func indexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	indexName := d.Get("name").(string)
	indexDefault := d.Get("default").(bool)

	o := d.Get("obj_name").([]interface{})[0].(map[string]interface{})

	b := materialize.NewIndexBuilder(
		meta.(*sqlx.DB),
		indexName,
		indexDefault,
		materialize.IdentifierSchemaStruct{
			Name:         o["name"].(string),
			SchemaName:   o["schema_name"].(string),
			DatabaseName: o["database_name"].(string),
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

	// set id
	i, err := materialize.IndexId(meta.(*sqlx.DB), indexName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return indexRead(ctx, d, meta)
}

func indexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	o := d.Get("obj_name").([]interface{})[0].(map[string]interface{})

	b := materialize.NewIndexBuilder(
		meta.(*sqlx.DB),
		d.Get("name").(string),
		d.Get("default").(bool),
		materialize.IdentifierSchemaStruct{
			Name:         o["name"].(string),
			SchemaName:   o["schema_name"].(string),
			DatabaseName: o["database_name"].(string),
		},
	)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
