package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var sinkKafkaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("sink", true, false),
	"schema_name":        SchemaNameSchema("sink", false),
	"database_name":      DatabaseNameSchema("sink", false),
	"qualified_sql_name": QualifiedNameSchema("sink"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("sink"),
	"size":               ObjectSizeSchema("sink"),
	"from": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "from",
		Description: "The name of the source, table or materialized view you want to send to the sink.",
		Required:    true,
		ForceNew:    false,
	}),
	"kafka_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "kafka_connection",
		Description: "The name of the Kafka connection to use in the sink.",
		Required:    true,
		ForceNew:    true,
	}),
	"topic": {
		Description: "The Kafka topic you want to subscribe to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"topic_replication_factor": {
		Description: "The replication factor to use when creating the Kafka topic (if the Kafka topic does not already exist).",
		Type:        schema.TypeInt,
		Optional:    true,
		ForceNew:    true,
	},
	"topic_partition_count": {
		Description: "The partition count to use when creating the Kafka topic (if the Kafka topic does not already exist).",
		Type:        schema.TypeInt,
		Optional:    true,
		ForceNew:    true,
	},
	"topic_config": {
		Description: "Any topic-level configs to use when creating the Kafka topic (if the Kafka topic does not already exist).",
		Type:        schema.TypeMap,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validateKafkaTopicConfigStringMap,
	},
	"compression_type": {
		Description:  "The type of compression to apply to messages before they are sent to Kafka.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(compressionTypes, true),
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
	"ownership_role": OwnershipRoleSchema(),
	"key_not_enforced": {
		Description: "Disable Materialize's validation of the key's uniqueness.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"headers": {
		Description: "The name of a column containing additional headers to add to each message emitted by the sink. The column must be of type map[text => text] or map[text => bytea].",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"partition_by": {
		Description: "A SQL expression used to partition the data in the Kafka sink. Can only be used with `ENVELOPE UPSERT`.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func SinkKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka sink establishes a link to a Kafka cluster that you want Materialize to write data to.",

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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SINK", Name: sinkName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSinkKafkaBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("from"); ok {
		from := materialize.GetIdentifierSchemaStruct(v)
		b.From(from)
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		b.Topic(v.(string))
	}

	if v, ok := d.GetOk("topic_replication_factor"); ok {
		b.TopicReplicationFactor(v.(int))
	}

	if v, ok := d.GetOk("topic_partition_count"); ok {
		b.TopicPartitionCount(v.(int))
	}

	if v, ok := d.GetOk("topic_config"); ok {
		config := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			config[k] = v.(string)
		}
		b.TopicConfig(config)
	}

	if v, ok := d.GetOk("compression_type"); ok {
		b.CompressionType(v.(string))
	}

	if v, ok := d.GetOk("key"); ok && len(v.([]interface{})) > 0 {
		keys, err := materialize.GetSliceValueString("key", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.Key(keys)
	}

	if v, ok := d.GetOk("key_not_enforced"); ok {
		b.KeyNotEnforced(v.(bool))
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

	if v, ok := d.GetOk("headers"); ok {
		b.Headers(v.(string))
	}

	if v, ok := d.GetOk("partition_by"); ok {
		b.PartitionBy(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	// set id
	i, err := materialize.SinkId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sinkRead(ctx, d, meta)
}
