package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	"item_name": {
		Description: "The name of the source, table or materialized view you want to send to the sink.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"kafka_connection": {
		Description: "The name of the Kafka connection to use in the source.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
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
	"schema_registry_connection": {
		Description: "The name of the connection to use for the shcema registry.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
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
		ReadContext:   SinkRead,
		UpdateContext: sinkKafkaUpdate,
		DeleteContext: sinkKafkaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sinkKafkaSchema,
	}
}

type SinkKafkaBuilder struct {
	sinkName                 string
	schemaName               string
	databaseName             string
	clusterName              string
	size                     string
	itemName                 string
	kafkaConnection          string
	topic                    string
	key                      []string
	format                   string
	envelope                 string
	schemaRegistryConnection string
	avroKeyFullname          string
	avroValueFullname        string
	snapshot                 bool
}

func (b *SinkKafkaBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.sinkName)
}

func newSinkKafkaBuilder(sinkName, schemaName, databaseName string) *SinkKafkaBuilder {
	return &SinkKafkaBuilder{
		sinkName:     sinkName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SinkKafkaBuilder) ClusterName(c string) *SinkKafkaBuilder {
	b.clusterName = c
	return b
}

func (b *SinkKafkaBuilder) Size(s string) *SinkKafkaBuilder {
	b.size = s
	return b
}

func (b *SinkKafkaBuilder) ItemName(i string) *SinkKafkaBuilder {
	b.itemName = i
	return b
}

func (b *SinkKafkaBuilder) KafkaConnection(k string) *SinkKafkaBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SinkKafkaBuilder) Topic(t string) *SinkKafkaBuilder {
	b.topic = t
	return b
}

func (b *SinkKafkaBuilder) Key(k []string) *SinkKafkaBuilder {
	b.key = k
	return b
}

func (b *SinkKafkaBuilder) Format(f string) *SinkKafkaBuilder {
	b.format = f
	return b
}

func (b *SinkKafkaBuilder) Envelope(e string) *SinkKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SinkKafkaBuilder) SchemaRegistryConnection(s string) *SinkKafkaBuilder {
	b.schemaRegistryConnection = s
	return b
}

func (b *SinkKafkaBuilder) AvroKeyFullname(a string) *SinkKafkaBuilder {
	b.avroKeyFullname = a
	return b
}

func (b *SinkKafkaBuilder) AvroValueFullname(a string) *SinkKafkaBuilder {
	b.avroValueFullname = a
	return b
}

func (b *SinkKafkaBuilder) Snapshot(s bool) *SinkKafkaBuilder {
	b.snapshot = s
	return b
}

func (b *SinkKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SINK %s`, b.qualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` FROM %s`, b.itemName))

	// Broker
	if b.kafkaConnection != "" {
		q.WriteString(fmt.Sprintf(` INTO KAFKA CONNECTION %s`, b.kafkaConnection))
	}

	if len(b.key) > 0 {
		o := strings.Join(b.key[:], ", ")
		q.WriteString(fmt.Sprintf(` KEY (%s)`, o))
	}

	if b.topic != "" {
		q.WriteString(fmt.Sprintf(` (TOPIC %s)`, QuoteString(b.topic)))
	}

	if b.format != "" {
		q.WriteString(fmt.Sprintf(` FORMAT %s`, b.format))
	}

	// CSR Options
	if b.schemaRegistryConnection != "" {
		q.WriteString(fmt.Sprintf(` USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.schemaRegistryConnection))
	}

	if b.avroKeyFullname != "" && b.avroValueFullname != "" {
		q.WriteString(fmt.Sprintf(` WITH (AVRO KEY FULLNAME %s AVRO VALUE FULLNAME %s)`, b.avroKeyFullname, b.avroValueFullname))
	}

	if b.envelope != "" {
		q.WriteString(fmt.Sprintf(` ENVELOPE %s`, b.envelope))
	}

	// With Options
	if b.size != "" || !b.snapshot {
		w := strings.Builder{}

		if b.size != "" {
			w.WriteString(fmt.Sprintf(` SIZE = %s`, QuoteString(b.size)))
		}

		if !b.snapshot {
			w.WriteString(` SNAPSHOT = false`)
		}

		q.WriteString(fmt.Sprintf(` WITH (%s)`, w.String()))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SinkKafkaBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SINK %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SinkKafkaBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SINK %s SET (SIZE = %s);`, b.qualifiedName(), QuoteString(newSize))
}

func (b *SinkKafkaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SINK %s;`, b.qualifiedName())
}

func (b *SinkKafkaBuilder) ReadId() string {
	return readSinkId(b.sinkName, b.schemaName, b.databaseName)
}

func sinkKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSinkKafkaBuilder(sinkName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("item_name"); ok {
		builder.ItemName(v.(string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		builder.KafkaConnection(v.(string))
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
		builder.SchemaRegistryConnection(v.(string))
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
	return SinkRead(ctx, d, meta)
}

func sinkKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSinkKafkaBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSinkKafkaBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return SinkRead(ctx, d, meta)
}

func sinkKafkaDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newSinkKafkaBuilder(sinkName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "sink"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
