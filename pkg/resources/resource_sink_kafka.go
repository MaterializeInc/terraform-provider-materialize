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

var sinkKafkaSchema = map[string]*schema.Schema{
	"name":               SchemaResourceName("sink", true, false),
	"schema_name":        SchemaResourceSchemaName("sink", false),
	"database_name":      SchemaResourceDatabaseName("sink", false),
	"qualified_sql_name": SchemaResourceQualifiedName("sink"),
	"sink_type":          SchemaResourceSinkType(),
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
		UpdateContext: sinkKafkaUpdate,
		DeleteContext: sinkKafkaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sinkKafkaSchema,
	}
}

func sinkKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewSinkKafkaBuilder(sinkName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("from"); ok {
		from := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.From(from)
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		builder.Topic(v.(string))
	}

	if v, ok := d.GetOk("key"); ok {
		builder.Key(v.([]string))
	}

	if v, ok := d.GetOk("format"); ok {
		format := materialize.GetSinkFormatSpecStruc(v)
		builder.Format(format)
	}

	if v, ok := d.GetOk("envelope"); ok {
		var envelope materialize.SinkEnvelopeStruct
		if v, ok := v.([]interface{})[0].(map[string]interface{})["upsert"]; ok {
			envelope.Upsert = v.(bool)
		}
		if v, ok := v.([]interface{})[0].(map[string]interface{})["debezium"]; ok {
			envelope.Debezium = v.(bool)
		}
		builder.Envelope(envelope)
	}

	if v, ok := d.GetOk("snapshot"); ok {
		builder.Snapshot(v.(bool))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "sink"); err != nil {
		return diag.FromErr(err)
	}
	return sinkRead(ctx, d, meta)
}

func sinkKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := materialize.NewSinkKafkaBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := materialize.NewSinkKafkaBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sinkRead(ctx, d, meta)
}

func sinkKafkaDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewSinkKafkaBuilder(sinkName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "sink"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
