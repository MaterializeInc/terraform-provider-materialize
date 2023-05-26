package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sinkKafkaSchema = map[string]*schema.Schema{
	"name":               NameSchema("sink", true, false),
	"schema_name":        SchemaNameSchema("sink", false),
	"database_name":      DatabaseNameSchema("sink", false),
	"qualified_sql_name": QualifiedNameSchema("sink"),
	"cluster_name": {
		Description:  "The cluster to maintain this sink. If not specified, the size option must be specified.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
	},
	"size": {
		Description:  "The size of the sink.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
		ValidateFunc: validation.StringInSlice(append(sourceSizes, localSizes...), true),
	},
	"from":             IdentifierSchema("from", "The name of the source, table or materialized view you want to send to the sink.", true),
	"kafka_connection": IdentifierSchema("kafka_connection", "The name of the Kafka connection to use in the sink.", true),
	"topic": {
		Description: "The Kafka topic you want to subscribe to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"key": {
		Description: "An optional list of columns to use for the Kafka key. If unspecified, the Kafka key is left unset.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
	},
	"format": SinkFormatSpecSchema("format", "How to decode raw bytes from different formats into data structures it can understand at runtime.", false),
	"envelope": {
		Description: "How to interpret records (e.g. Debezium, Upsert).",
		Type:        schema.TypeList,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"upsert": {
					Description:   "The sink emits data with upsert semantics: updates and inserts for the given key are expressed as a value, and deletes are expressed as a null value payload in Kafka.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.debezium"},
				},
				"debezium": {
					Description:   "The generated schemas have a Debezium-style diff envelope to capture changes in the input view or source.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.upsert"},
				},
			},
		},
		Optional: true,
		ForceNew: true,
	},
	"snapshot": {
		Description: "Whether to emit the consolidated results of the query before the sink was created at the start of the sink.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     true,
	},
}

func SinkKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A sink describes an external system you want Materialize to write data to, and provides details about how to encode that data.",

		CreateContext: sinkKafkaCreate,
		ReadContext:   sinkRead,
		UpdateContext: sinkUpdate,
		DeleteContext: sinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sinkKafkaSchema,
	}
}

func sinkKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewSinkKafkaBuilder(meta.(*sqlx.DB), sinkName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
	}

	if v, ok := d.GetOk("from"); ok {
		from := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.From(from)
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		b.Topic(v.(string))
	}

	if v, ok := d.GetOk("key"); ok {
		keys := materialize.GetSliceValueString(v.([]interface{}))
		b.Key(keys)
	}

	if v, ok := d.GetOk("format"); ok {
		format := materialize.GetSinkFormatSpecStruc(v)
		b.Format(format)
	}

	if v, ok := d.GetOk("envelope"); ok {
		envelope := materialize.GetSinkKafkaEnelopeStruct(v)
		b.Envelope(envelope)
	}

	if v, ok := d.GetOk("snapshot"); ok {
		b.Snapshot(v.(bool))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.SinkId(meta.(*sqlx.DB), sinkName, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return sinkRead(ctx, d, meta)
}
