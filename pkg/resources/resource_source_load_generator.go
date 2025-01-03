package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var tick_interval = &schema.Schema{
	Description: "The interval at which the next datum should be emitted. Defaults to one second.",
	Type:        schema.TypeString,
	Optional:    true,
	ForceNew:    true,
}

var scale_factor = &schema.Schema{
	Description: "The scale factor for the generator. Defaults to 0.01 (~ 10MB).",
	Type:        schema.TypeFloat,
	Optional:    true,
	Default:     0.01,
	ForceNew:    true,
}

var scale_factor_deprecated = &schema.Schema{
	Description: "(Deprecated) The scale factor for the generator. Defaults to 0.01 (~ 10MB).",
	Deprecated:  "Scale factor is deprecated and will be removed in a future release.",
	Type:        schema.TypeFloat,
	Optional:    true,
	Default:     0.01,
	ForceNew:    true,
}

var sourceLoadgenSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("source"),
	"size":               ObjectSizeSchema("source"),
	"load_generator_type": {
		Description:  fmt.Sprintf("The load generator types: %s.", loadGeneratorTypes),
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(loadGeneratorTypes, true),
	},
	"counter_options": {
		Description: "Counter Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor_deprecated,
				"max_cardinality": {
					Description: "Causes the generator to delete old values to keep the collection at most a given size. Defaults to unlimited.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"auction_options", "marketing_options", "tpch_options"},
	},
	"auction_options": {
		Description: "Auction Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor_deprecated,
			},
		},
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"counter_options", "marketing_options", "tpch_options"},
	},
	"marketing_options": {
		Description: "Marketing Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor_deprecated,
			},
		},
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"counter_options", "auction_options", "tpch_options"},
	},
	"tpch_options": {
		Description: "TPCH Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor,
			},
		},
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"counter_options", "auction_options", "marketing_options"},
	},
	"key_value_options": {
		Description: "KEY VALUE Load Generator Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"keys": {
					Description: "The number of keys in the source. This must be divisible by the product of 'partitions' and 'batch_size'.",
					Type:        schema.TypeInt,
					Required:    true,
					ForceNew:    true,
				},
				"snapshot_rounds": {
					Description: "The number of rounds of data to produce as the source starts up.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
				"transactional_snapshot": {
					Description: "Whether to emit the snapshot as a singular transaction.",
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
				},
				"value_size": {
					Description: "The number of bytes in each value.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
				"tick_interval": tick_interval,
				"seed": {
					Description: "A per-source seed for seeding the random data.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
				"partitions": {
					Description: "The number of partitions to spread the keys across.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
				"batch_size": {
					Description: "The number of keys per partition to produce in each update.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		ForceNew:      true,
		ConflictsWith: []string{"counter_options", "auction_options", "marketing_options", "tpch_options"},
	},
	"all_tables": {
		Description:   "Whether to include all tables in the source. Compatible with `auction_options`, `marketing_options`, and `tpch_options`. If not specified, use the `materialize_source_table_load_generator` resource to specify tables to include.",
		Type:          schema.TypeBool,
		Optional:      true,
		Default:       false,
		ConflictsWith: []string{"counter_options", "key_value_options"},
		ForceNew:      true,
	},
	"expose_progress": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "expose_progress",
		Description: "The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`.",
		Required:    false,
		ForceNew:    true,
	}),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceLoadgen() *schema.Resource {
	return &schema.Resource{
		Description: "A load generator source produces synthetic data for use in demos and performance tests.",

		CreateContext: sourceLoadgenCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceUpdate,
		DeleteContext: sourceLoadgenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceLoadgenSchema,
	}
}

func sourceLoadgenCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceLoadgenBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("expose_progress"); ok {
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
	}

	if v, ok := d.GetOk("load_generator_type"); ok {
		b.LoadGeneratorType(v.(string))
	}

	if v, ok := d.GetOk("counter_options"); ok {
		o := materialize.GetCounterOptionsStruct(v)
		b.CounterOptions(o)
	}

	if v, ok := d.GetOk("auction_options"); ok {
		o := materialize.GetAuctionOptionsStruct(v)
		b.AuctionOptions(o)
	}

	if v, ok := d.GetOk("marketing_options"); ok {
		o := materialize.GetMarketingOptionsStruct(v)
		b.MarketingOptions(o)
	}

	if v, ok := d.GetOk("tpch_options"); ok {
		o := materialize.GetTPCHOptionsStruct(v)
		b.TPCHOptions(o)
	}

	if v, ok := d.GetOk("key_value_options"); ok {
		o := materialize.GetKeyValueOptionsStruct(v)
		b.KeyValueOptions(o)
	}

	// all_tables
	if v, ok := d.GetOk("all_tables"); ok && v.(bool) {
		b.AllTables()
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceRead(ctx, d, meta)
}

func sourceLoadgenDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if err := b.DropCascade(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
