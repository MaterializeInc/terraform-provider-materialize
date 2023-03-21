package resources

import (
	"context"
	"log"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sinkKafkaSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the sink.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the sink schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
	},
	"database_name": {
		Description: "The identifier for the sink database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the sink.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"sink_type": {
		Description: "The type of sink.",
		Type:        schema.TypeString,
		Computed:    true,
	},
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
	"from":             IdentifierSchema("from", "The name of the source, table or materialized view you want to send to the sink.", true, false),
	"kafka_connection": IdentifierSchema("kafka_connection", "The name of the Kafka connection to use in the sink.", true, false),
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
	"format": {
		Description: "How to decode raw bytes from different formats into data structures it can understand at runtime.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"envelope": {
		Description:  "How to interpret records (e.g. Append Only, Upsert).",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(envelopes, true),
	},
	"schema_registry_connection": IdentifierSchema("schema_registry_connection", "The name of the connection to use for the shcema registry.", false, true),
	"avro_key_fullname": {
		Description:  "ets the Avro fullname on the generated key schema, if a KEY is specified. When used, a value must be specified for AVRO VALUE FULLNAME.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"avro_key_fullname", "avro_value_fullname"},
	},
	"avro_value_fullname": {
		Description:  "Sets the Avro fullname on the generated value schema. When KEY is specified, AVRO KEY FULLNAME must additionally be specified.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"avro_key_fullname", "avro_value_fullname"},
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
		builder.Format(v.(string))
	}

	if v, ok := d.GetOk("envelope"); ok {
		builder.Envelope(v.(string))
	}

	if v, ok := d.GetOk("schema_registry_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.SchemaRegistryConnection(conn)
	}

	if v, ok := d.GetOk("avro_key_fullname"); ok {
		builder.AvroKeyFullname(v.(string))
	}

	if v, ok := d.GetOk("avro_value_fullname"); ok {
		builder.AvroValueFullname(v.(string))
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
