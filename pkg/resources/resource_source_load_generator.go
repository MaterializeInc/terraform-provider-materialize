package resources

import (
	"context"
	"fmt"
	"log"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceLoadgenSchema = map[string]*schema.Schema{
	"name":           SchemaResourceName("source", true, false),
	"schema_name":    SchemaResourceSchemaName("source", false),
	"database_name":  SchemaResourceDatabaseName("source", false),
	"qualified_name": SchemaResourceQualifiedName("source"),
	"source_type":    SchemaResourceSourceType(),
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
	"tick_interval": {
		Description: "The interval at which the next datum should be emitted. Defaults to one second.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"scale_factor": {
		Description: "The scale factor for the TPCH generator. Defaults to 0.01 (~ 10MB).",
		Type:        schema.TypeFloat,
		Optional:    true,
		Default:     0.01,
		ForceNew:    true,
	},
	"max_cardinality": {
		Description: "Valid for the COUNTER generator. Causes the generator to delete old values to keep the collection at most a given size. Defaults to unlimited.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"table": {
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

	if v, ok := d.GetOk("tick_interval"); ok {
		builder.TickInterval(v.(string))
	}

	if v, ok := d.GetOk("scale_factor"); ok {
		builder.ScaleFactor(v.(float64))
	}

	if v, ok := d.GetOk("max_cardinality"); ok {
		builder.MaxCardinality(v.(bool))
	}

	if v, ok := d.GetOk("table"); ok {
		var tables []materialize.TableLoadgen
		for _, table := range v.([]interface{}) {
			t := table.(map[string]interface{})
			tables = append(tables, materialize.TableLoadgen{
				Name:  t["name"].(string),
				Alias: t["alias"].(string),
			})
		}
		builder.Table(tables)
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
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := materialize.NewSourceLoadgenBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := materialize.NewSourceLoadgenBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
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
