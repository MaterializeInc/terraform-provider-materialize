package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-materialize-provider/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceKafkaSchema = map[string]*schema.Schema{
	"name":               SchemaResourceName("source", true, false),
	"schema_name":        SchemaResourceSchemaName("source", false),
	"database_name":      SchemaResourceDatabaseName("source", false),
	"qualified_sql_name": SchemaResourceQualifiedName("source"),
	"source_type":        SchemaResourceSourceType(),
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
	"primary_key": {
		Description: "Declare a set of columns as a primary key.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
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
}

func SourceKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourceKafkaCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceKafkaUpdate,
		DeleteContext: sourceKafkaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceKafkaSchema,
	}
}

func sourceKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSourceKafkaBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		builder.Topic(v.(string))
	}

	if v, ok := d.GetOk("include_key"); ok && v.(bool) {
		builder.IncludeKey()
	}

	if v, ok := d.GetOk("include_headers"); ok && v.(bool) {
		builder.IncludeHeaders()
	}

	if v, ok := d.GetOk("include_partition"); ok && v.(bool) {
		builder.IncludePartition()
	}

	if v, ok := d.GetOk("include_offset"); ok && v.(bool) {
		builder.IncludeOffset()
	}

	if v, ok := d.GetOk("include_timestamp"); ok && v.(bool) {
		builder.IncludeTimestamp()
	}

	if v, ok := d.GetOk("format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		builder.Format(format)
	}

	if v, ok := d.GetOk("key_format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		builder.KeyFormat(format)
	}

	if v, ok := d.GetOk("value_format"); ok {
		format := materialize.GetFormatSpecStruc(v)
		builder.ValueFormat(format)
	}

	if v, ok := d.GetOk("envelope"); ok {
		var envelope materialize.KafkaSourceEnvelopeStruct
		if v, ok := v.([]interface{})[0].(map[string]interface{})["upsert"]; ok {
			envelope.Upsert = v.(bool)
		}
		if v, ok := v.([]interface{})[0].(map[string]interface{})["debezium"]; ok {
			envelope.Debezium = v.(bool)
		}
		if v, ok := v.([]interface{})[0].(map[string]interface{})["none"]; ok {
			envelope.None = v.(bool)
		}
		builder.Envelope(envelope)
	}

	if v, ok := d.GetOk("primary_key"); ok {
		builder.PrimaryKey(v.([]string))
	}

	if v, ok := d.GetOk("start_offset"); ok {
		builder.StartOffset(v.([]int))
	}

	if v, ok := d.GetOk("start_timestamp"); ok {
		builder.StartTimestamp(v.(int))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "source"); err != nil {
		return diag.FromErr(err)
	}
	return sourceRead(ctx, d, meta)
}

func sourceKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := materialize.NewSourceKafkaBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := materialize.NewSourceKafkaBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sourceRead(ctx, d, meta)
}

func sourceKafkaDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewSourceKafkaBuilder(sourceName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "source"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
