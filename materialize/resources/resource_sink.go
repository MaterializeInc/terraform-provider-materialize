package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sinkSchema = map[string]*schema.Schema{
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
	"cluster_name": {
		Description:   "The cluster to maintain this sink. If not specified, the size option must be specified.",
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"size"},
	},
	"size": {
		Description:   "The size of the sink.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ValidateFunc:  validation.StringInSlice(append(sourceSizes, localSizes...), true),
		ConflictsWith: []string{"cluster_name"},
	},
	"item_name": {
		Description: "The name of the source, table or materialized view you want to send to the sink.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	// Broker
	"kafka_connection": {
		Description:  "The name of the Kafka connection to use in the source.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"kafka_connection", "topic"},
	},
	"topic": {
		Description:  "The Kafka topic you want to subscribe to.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"kafka_connection", "topic"},
	},
	"key": {
		Description: "An optional list of columns to use for the Kafka key. If unspecified, the Kafka key is left unset.",
		Type:        schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional: true,
		ForceNew: true,
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
	// "snapshot": {
	// 	Description: "Whether to emit the consolidated results of the query before the sink was created at the start of the sink.",
	// 	Type:        schema.TypeBool,
	// 	Optional:    true,
	// 	ForceNew:    true,
	// 	Default:     true,
	// },
}

func Sink() *schema.Resource {
	return &schema.Resource{
		Description: "A sink describes an external system you want Materialize to write data to, and provides details about how to encode that data.",

		CreateContext: sinkCreate,
		ReadContext:   sinkRead,
		UpdateContext: sinkUpdate,
		DeleteContext: sinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sinkSchema,
	}
}

type SinkBuilder struct {
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
	// snapshot                 bool
}

func newSinkBuilder(sinkName, schemaName, databaseName string) *SinkBuilder {
	return &SinkBuilder{
		sinkName:     sinkName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SinkBuilder) ClusterName(c string) *SinkBuilder {
	b.clusterName = c
	return b
}

func (b *SinkBuilder) Size(s string) *SinkBuilder {
	b.size = s
	return b
}

func (b *SinkBuilder) ItemName(i string) *SinkBuilder {
	b.itemName = i
	return b
}

func (b *SinkBuilder) KafkaConnection(k string) *SinkBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SinkBuilder) Topic(t string) *SinkBuilder {
	b.topic = t
	return b
}

func (b *SinkBuilder) Key(k []string) *SinkBuilder {
	b.key = k
	return b
}

func (b *SinkBuilder) Format(f string) *SinkBuilder {
	b.format = f
	return b
}

func (b *SinkBuilder) Envelope(e string) *SinkBuilder {
	b.envelope = e
	return b
}

func (b *SinkBuilder) SchemaRegistryConnection(s string) *SinkBuilder {
	b.schemaRegistryConnection = s
	return b
}

func (b *SinkBuilder) AvroKeyFullname(a string) *SinkBuilder {
	b.avroKeyFullname = a
	return b
}

func (b *SinkBuilder) AvroValueFullname(a string) *SinkBuilder {
	b.avroValueFullname = a
	return b
}

// func (b *SinkBuilder) Snapshot(s bool) *SinkBuilder {
// 	b.snapshot = s
// 	return b
// }

func (b *SinkBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SINK %s.%s.%s`, b.databaseName, b.schemaName, b.sinkName))

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
		q.WriteString(fmt.Sprintf(` (TOPIC '%s')`, b.topic))
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
	if b.size != "" {
		w := strings.Builder{}

		if b.size != "" {
			w.WriteString(fmt.Sprintf(`SIZE = '%s'`, b.size))
		}

		// if !b.snapshot {
		// 	w.WriteString(`SNAPSHOT = false`)
		// }

		q.WriteString(fmt.Sprintf(` WITH (%s)`, w.String()))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SinkBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_sinks.id
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.sinkName, b.schemaName, b.databaseName)
}

func (b *SinkBuilder) Rename(newName string) string {
	return fmt.Sprintf(`ALTER SINK %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.sinkName, b.databaseName, b.schemaName, newName)
}

func (b *SinkBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SINK %s.%s.%s SET (SIZE = '%s');`, b.databaseName, b.schemaName, b.sinkName, newSize)
}

func (b *SinkBuilder) Drop() string {
	return fmt.Sprintf(`DROP SINK %s.%s.%s;`, b.databaseName, b.schemaName, b.sinkName)
}

func readSinkParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.type,
			mz_sinks.size,
			mz_sinks.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.id = '%s';`, id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _sink struct {
	name            sql.NullString `db:"name"`
	schema_name     sql.NullString `db:"schema_name"`
	database_name   sql.NullString `db:"database_name"`
	sink_type       sql.NullString `db:"sink_type"`
	size            sql.NullString `db:"size"`
	envelope_type   sql.NullString `db:"envelope_type"`
	connection_name sql.NullString `db:"connection_name"`
	cluster_name    sql.NullString `db:"cluster_name"`
}

func sinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSinkParams(i)

	readResource(conn, d, i, q, _sink{}, "sink")
	return nil
}

func sinkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSinkBuilder(sinkName, schemaName, databaseName)

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

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "sink")
	return sourceRead(ctx, d, meta)
}

func sinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSinkBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSinkBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sinkRead(ctx, d, meta)
}

func sinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSinkBuilder(sinkName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "sink")
	return nil
}
