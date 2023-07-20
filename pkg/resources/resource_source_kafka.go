package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceKafkaSchema = map[string]*schema.Schema{
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
		ForceNew:     true,
	},
	"size": {
		Description:  "The size of the source.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
		ValidateFunc: validation.StringInSlice(append(sourceSizes, localSizes...), true),
	},
	"kafka_connection": IdentifierSchema("kafka_connection", "The Kafka connection to use in the source.", true),
	"topic": {
		Description: "The Kafka topic you want to subscribe to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"include_key": {
		Description: "Include a column containing the Kafka message key. If the key is encoded using a format that includes schemas, the column will take its name from the schema. For unnamed formats (e.g. TEXT), the column will be named \"key\".",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_headers": {
		Description: "Include message headers.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"include_partition": {
		Description: "Include a partition column containing the Kafka message partition",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"include_offset": {
		Description: "Include an offset column containing the Kafka message offset.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"include_timestamp": {
		Description: "Include a timestamp column containing the Kafka message timestamp.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"format":       FormatSpecSchema("format", "How to decode raw bytes from different formats into data structures Materialize can understand at runtime.", false),
	"key_format":   FormatSpecSchema("key_format", "Set the key format explicitly.", false),
	"value_format": FormatSpecSchema("value_format", "Set the value format explicitly.", false),
	"envelope": {
		Description: "How Materialize should interpret records (e.g. append-only, upsert)..",
		Type:        schema.TypeList,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"upsert": {
					Description:   "Use the upsert envelope, which uses message keys to handle CRUD operations.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.debezium", "envelope.0.none"},
				},
				"debezium": {
					Description:   "Use the Debezium envelope, which uses a diff envelope to handle CRUD operations.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.upsert", "envelope.0.none"},
				},
				"none": {
					Description:   "Use an append-only envelope. This means that records will only be appended and cannot be updated or deleted.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.upsert", "envelope.0.debezium"},
				},
			},
		},
		Optional: true,
		ForceNew: true,
	},
	"start_offset": {
		Description: "Read partitions from the specified offset.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeInt},
		Optional:    true,
		ForceNew:    true,
	},
	"start_timestamp": {
		Description: "Use the specified value to set \"START OFFSET\" based on the Kafka timestamp.",
		Type:        schema.TypeInt,
		Optional:    true,
		ForceNew:    true,
	},
	"expose_progress": {
		Description: "The name of the progress subsource for the source. If this is not specified, the subsource will be named `<src_name>_progress`.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"ownership_role": OwnershipRole(),
}

func SourceKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka source describes a Kafka cluster you want Materialize to read data from.",

		CreateContext: sourceKafkaCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceUpdate,
		DeleteContext: sourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceKafkaSchema,
	}
}

func sourceKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.ObjectSchemaStruct{Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceKafkaBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		b.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		b.Topic(v.(string))
	}

	if v, ok := d.GetOk("include_key"); ok && v.(bool) {
		b.IncludeKey()
	}

	if v, ok := d.GetOk("include_headers"); ok && v.(bool) {
		b.IncludeHeaders()
	}

	if v, ok := d.GetOk("include_partition"); ok && v.(bool) {
		b.IncludePartition()
	}

	if v, ok := d.GetOk("include_offset"); ok && v.(bool) {
		b.IncludeOffset()
	}

	if v, ok := d.GetOk("include_timestamp"); ok && v.(bool) {
		b.IncludeTimestamp()
	}

	if v, ok := d.GetOk("format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		b.Format(format)
	}

	if v, ok := d.GetOk("key_format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		b.KeyFormat(format)
	}

	if v, ok := d.GetOk("value_format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		b.ValueFormat(format)
	}

	if v, ok := d.GetOk("envelope"); ok {
		envelope := materialize.GetSourceKafkaEnelopeStruct(v)
		b.Envelope(envelope)
	}

	if v, ok := d.GetOk("start_offset"); ok {
		so := materialize.GetSliceValueInt(v.([]interface{}))
		b.StartOffset(so)
	}

	if v, ok := d.GetOk("start_timestamp"); ok {
		b.StartTimestamp(v.(int))
	}

	if v, ok := d.GetOk("expose_progress"); ok {
		b.ExposeProgress(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "SOURCE", o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SourceId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return sourceRead(ctx, d, meta)
}
