package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
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

var table = &schema.Schema{
	Description: "Creates subsources for specific tables.",
	Type:        schema.TypeList,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the table.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"alias": {
				Description: "The alias of the table.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	},
	Optional: true,
	MinItems: 1,
	ForceNew: true,
}

var sourceLoadgenSchema = map[string]*schema.Schema{
	"name":               NameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"cluster_name": {
		Description:  "The cluster to maintain this source. If not specified, the size option must be specified.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
	},
	"size": {
		Description:  "The size of the source.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
		ValidateFunc: validation.StringInSlice(append(sourceSizes, localSizes...), true),
	},
	"load_generator_type": {
		Description:  fmt.Sprintf("The load generator types: %s.", loadGeneratorTypes),
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(loadGeneratorTypes, true),
	},
	"counter_options": {
		Description: "Counter Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor,
				"max_cardinality": {
					Description: "Causes the generator to delete old values to keep the collection at most a given size. Defaults to unlimited.",
					Type:        schema.TypeInt,
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
		Optional:     true,
		MinItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"counter_options", "auction_options", "tpch_options"},
	},
	"auction_options": {
		Description: "Auction Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor,
				"table":         table,
			},
		},
		Optional:     true,
		MinItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"counter_options", "auction_options", "tpch_options"},
	},
	"tpch_options": {
		Description: "TPCH Options.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_interval": tick_interval,
				"scale_factor":  scale_factor,
				"table":         table,
			},
		},
		Optional:     true,
		MinItems:     1,
		ForceNew:     true,
		ExactlyOneOf: []string{"counter_options", "auction_options", "tpch_options"},
	},
}

func SourceLoadgen() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourceLoadgenCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceLoadgenUpdate,
		DeleteContext: sourceLoadgenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceLoadgenSchema,
	}
}

func sourceLoadgenCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSourceLoadgenBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("load_generator_type"); ok {
		builder.LoadGeneratorType(v.(string))
	}

	if v, ok := d.GetOk("counter_options"); ok {
		o := materialize.GetCounterOptionsStruct(v)
		builder.CounterOptions(o)
	}

	if v, ok := d.GetOk("auction_options"); ok {
		o := materialize.GetAuctionOptionsStruct(v)
		builder.AuctionOptions(o)
	}

	if v, ok := d.GetOk("tpch_options"); ok {
		o := materialize.GetTPCHOptionsStruct(v)
		builder.TPCHOptions(o)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "source"); err != nil {
		return diag.FromErr(err)
	}
	return sourceRead(ctx, d, meta)
}

func sourceLoadgenUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("size") {
		_, newSize := d.GetChange("size")

		q := materialize.NewSourceLoadgenBuilder(sourceName, schemaName, databaseName).UpdateSize(newSize.(string))

		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not resize sink: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := materialize.NewSourceLoadgenBuilder(sourceName, schemaName, databaseName).Rename(newName.(string))

		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename sink: %s", q)
			return diag.FromErr(err)
		}
	}

	return sourceRead(ctx, d, meta)
}

func sourceLoadgenDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewSourceLoadgenBuilder(sourceName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "source"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
