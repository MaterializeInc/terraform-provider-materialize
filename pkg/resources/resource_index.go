package resources

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description:  "The identifier for the index.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
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
		Description:  "Creates a default index using all inferred columns are used. Required if col_expr is not set.",
		Type:         schema.TypeBool,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"qualified_sql_name": QualifiedNameSchema("index"),
	"comment":            CommentSchema(false),
	"obj_name": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "obj_name",
		Description: "The name of the source, view, or materialized view on which you want to create an index.",
		Required:    true,
		ForceNew:    true,
	}),
	"cluster_name": {
		Description: "The cluster to maintain this index.",
		Type:        schema.TypeString,
		Required:    true,
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
					ForceNew:    true,
				},
			},
		},
		Optional:      true,
		ConflictsWith: []string{"default"},
		Computed:      true,
		ForceNew:      true,
	},
	"region": RegionSchema(),
}

func Index() *schema.Resource {
	return &schema.Resource{
		Description: "Indexes represent query results stored in memory. Recent Materialize versions do not drop an index that a materialized view or another index still reads from, and Terraform cannot see that link, so replacing such an index fails. Add `lifecycle { replace_triggered_by = [materialize_index.<name>.id] }` to the resources that read from it so Terraform replaces them together with the index. Anything built on those resources, such as their own indexes, needs the same block pointing at them.",

		CreateContext: indexCreate,
		ReadContext:   indexRead,
		UpdateContext: indexUpdate,
		DeleteContext: indexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: indexSchema,
	}
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanIndex(metaDb, utils.ExtractId(i))
	if errors.Is(err, sql.ErrNoRows) {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

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

	// Get and set the index columns
	indexColumns, err := materialize.ListIndexColumns(metaDb, utils.ExtractId(i))
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

	colExpr := d.Get("col_expr").([]interface{})

	if !indexDefault && len(colExpr) == 0 {
		return diag.Errorf("col_expr is required when creating a non-default index")
	}

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Index, Name: indexName}
	b := materialize.NewIndexBuilder(
		metaDb,
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
			return rollbackCreate(b, o, "comment", err)
		}
	}

	// set id
	if indexDefault {
		// For default indexes, find the index by the object it's on
		idxParams, err := materialize.FindDefaultIndexByObject(
			metaDb,
			obj["name"].(string),
			obj["schema_name"].(string),
			obj["database_name"].(string),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		// Set the real generated name in the state
		if err := d.Set("name", idxParams.IndexName.String); err != nil {
			return diag.FromErr(err)
		}

		// Set the correct ID
		d.SetId(utils.TransformIdWithRegion(string(region), idxParams.IndexId.String))
	} else {
		// Original code for named indexes
		i, err := materialize.IndexId(metaDb, indexName)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(utils.TransformIdWithRegion(string(region), i))
	}

	return indexRead(ctx, d, meta)
}

func indexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	indexName := d.Get("name").(string)
	o := materialize.MaterializeObject{ObjectType: materialize.Index, Name: indexName}

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return indexRead(ctx, d, meta)
}

func indexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	obj := d.Get("obj_name").([]interface{})[0].(map[string]interface{})
	name := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Index, Name: name}
	b := materialize.NewIndexBuilder(
		metaDb,
		o,
		d.Get("default").(bool),
		materialize.IdentifierSchemaStruct{
			Name:         obj["name"].(string),
			SchemaName:   obj["schema_name"].(string),
			DatabaseName: obj["database_name"].(string),
		},
	)

	if err := dropIndex(ctx, b.Drop); err != nil {
		if materialize.IsDependentObjectsStillExist(err) {
			return diag.Diagnostics{{Severity: diag.Error, Summary: err.Error(), Detail: indexReadersHint}}
		}
		return diag.FromErr(err)
	}
	return nil
}

// Terraform cannot see that a materialized view reads from an index, so when
// the index is replaced it drops the index without touching the view. The view
// has to be told to follow the index.
const indexReadersHint = "The objects named in the error still read from this index, so Materialize cannot drop it. " +
	"If the index is being replaced, add `lifecycle { replace_triggered_by = [materialize_index.<name>.id] }` " +
	"to each of those resources so Terraform recreates them together with the index."

// Materialize refuses to drop an index while a materialized view or another
// index reads from it. terraform destroy drops those readers in parallel with
// the index, because the configuration does not link them, so the index drop
// gets a moment for them to go before it fails.
var indexDropRetryTimeout = time.Minute

func dropIndex(ctx context.Context, drop func() error) error {
	return retry.RetryContext(ctx, indexDropRetryTimeout, func() *retry.RetryError {
		err := drop()
		if err == nil {
			return nil
		}
		if materialize.IsDependentObjectsStillExist(err) {
			log.Printf("[INFO] index is still read by other objects, retrying drop: %s", err)
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})
}
