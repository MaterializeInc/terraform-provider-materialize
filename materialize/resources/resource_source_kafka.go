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

var sourceKafkaSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the source.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the source schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
	},
	"database_name": {
		Description: "The identifier for the source database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"source_type": {
		Description: "The type of source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
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
	"include_key": {
		Description: "Include a column containing the Kafka message key. If the key is encoded using a format that includes schemas, the column will take its name from the schema. For unnamed formats (e.g. TEXT), the column will be named \"key\".",
		Type:        schema.TypeString,
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
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_offset": {
		Description: "Include an offset column containing the Kafka message offset.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_timestamp": {
		Description: "Include a timestamp column containing the Kafka message timestamp.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"format": {
		Description: "How to decode raw bytes from different formats into data structures Materialize can understand at runtime.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"key_format": {
		Description: "Set the key and value encodings explicitly.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"envelope": {
		Description:  "How Materialize should interpret records (e.g. append-only, upsert).",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(envelopes, true),
	},
	"schema_registry_connection": {
		Description: "The name of a schema registry connection.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"key_strategy": {
		Description:  "How Materialize will define the Avro schema reader key strategy.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(strategy, true),
	},
	"value_strategy": {
		Description:  "How Materialize will define the Avro schema reader value strategy.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(strategy, true),
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
		ReadContext:   SourceRead,
		UpdateContext: sourceKafkaUpdate,
		DeleteContext: sourceKafkaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceKafkaSchema,
	}
}

type SourceKafkaBuilder struct {
	sourceName               string
	schemaName               string
	databaseName             string
	clusterName              string
	size                     string
	kafkaConnection          string
	topic                    string
	includeKey               string
	includeHeaders           bool
	includePartition         string
	includeOffset            string
	includeTimestamp         string
	format                   string
	keyFormat                string
	envelope                 string
	schemaRegistryConnection string
	keyStrategy              string
	valueStrategy            string
	primaryKey               []string
	startOffset              []int
	startTimestamp           int
}

func (b *SourceKafkaBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.sourceName)
}

func newSourceKafkaBuilder(sourceName, schemaName, databaseName string) *SourceKafkaBuilder {
	return &SourceKafkaBuilder{
		sourceName:   sourceName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SourceKafkaBuilder) ClusterName(c string) *SourceKafkaBuilder {
	b.clusterName = c
	return b
}

func (b *SourceKafkaBuilder) Size(s string) *SourceKafkaBuilder {
	b.size = s
	return b
}

func (b *SourceKafkaBuilder) KafkaConnection(k string) *SourceKafkaBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SourceKafkaBuilder) Topic(t string) *SourceKafkaBuilder {
	b.topic = t
	return b
}

func (b *SourceKafkaBuilder) IncludeKey(i string) *SourceKafkaBuilder {
	b.includeKey = i
	return b
}

func (b *SourceKafkaBuilder) IncludeHeaders() *SourceKafkaBuilder {
	b.includeHeaders = true
	return b
}

func (b *SourceKafkaBuilder) IncludePartition(i string) *SourceKafkaBuilder {
	b.includePartition = i
	return b
}

func (b *SourceKafkaBuilder) IncludeOffset(i string) *SourceKafkaBuilder {
	b.includeOffset = i
	return b
}

func (b *SourceKafkaBuilder) IncludeTimestamp(i string) *SourceKafkaBuilder {
	b.includeTimestamp = i
	return b
}

func (b *SourceKafkaBuilder) Format(f string) *SourceKafkaBuilder {
	b.format = f
	return b
}

func (b *SourceKafkaBuilder) Envelope(e string) *SourceKafkaBuilder {
	b.envelope = e
	return b
}

func (b *SourceKafkaBuilder) SchemaRegistryConnection(s string) *SourceKafkaBuilder {
	b.schemaRegistryConnection = s
	return b
}

func (b *SourceKafkaBuilder) KeyFormat(k string) *SourceKafkaBuilder {
	b.keyFormat = k
	return b
}

func (b *SourceKafkaBuilder) KeyStrategy(k string) *SourceKafkaBuilder {
	b.keyStrategy = k
	return b
}

func (b *SourceKafkaBuilder) ValueStrategy(v string) *SourceKafkaBuilder {
	b.valueStrategy = v
	return b
}

func (b *SourceKafkaBuilder) PrimaryKey(p []string) *SourceKafkaBuilder {
	b.primaryKey = p
	return b
}

func (b *SourceKafkaBuilder) StartOffset(s []int) *SourceKafkaBuilder {
	b.startOffset = s
	return b
}

func (b *SourceKafkaBuilder) StartTimestamp(s int) *SourceKafkaBuilder {
	b.startTimestamp = s
	return b
}

func (b *SourceKafkaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.qualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` FROM KAFKA CONNECTION %s (TOPIC '%s')`, b.kafkaConnection, b.topic))

	// Format
	if b.keyFormat != "" {
		q.WriteString(fmt.Sprintf(` KEY FORMAT %s VALUE FORMAT %s`, b.keyFormat, b.format))
	} else {
		q.WriteString(fmt.Sprintf(` FORMAT %s`, b.format))
	}

	if b.schemaRegistryConnection != "" {
		q.WriteString(fmt.Sprintf(` USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.schemaRegistryConnection))
	}

	// Key Constraint
	if len(b.primaryKey) > 0 {
		k := strings.Join(b.primaryKey[:], ", ")
		q.WriteString(fmt.Sprintf(` PRIMARY KEY (%s) NOT ENFORCED`, k))
	}

	// Time-based Offsets
	if len(b.startOffset) > 0 {
		k := strings.Join(strings.Fields(fmt.Sprint(b.startOffset)), ", ")
		q.WriteString(fmt.Sprintf(` START OFFSET %s`, k))
	}

	if b.startTimestamp != 0 {
		q.WriteString(fmt.Sprintf(` START TIMESTAMP %d`, b.startTimestamp))
	}

	// Strategy
	if b.keyStrategy != "" {
		q.WriteString(fmt.Sprintf(` KEY STRATEGY %s`, b.keyStrategy))
	}

	if b.valueStrategy != "" {
		q.WriteString(fmt.Sprintf(` VALUE STRATEGY %s`, b.valueStrategy))
	}

	// Metadata
	var i []string

	if b.includeKey != "" {
		i = append(i, b.includeKey)
	}

	if b.includeHeaders {
		i = append(i, "HEADERS")
	}

	if b.includePartition != "" {
		i = append(i, b.includePartition)
	}

	if b.includeOffset != "" {
		i = append(i, b.includeOffset)
	}

	if b.includeTimestamp != "" {
		i = append(i, b.includeTimestamp)
	}

	if len(i) > 0 {
		o := strings.Join(i[:], ", ")
		q.WriteString(fmt.Sprintf(` INCLUDE %s`, o))
	}

	if b.envelope != "" {
		q.WriteString(fmt.Sprintf(` ENVELOPE %s`, b.envelope))
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = '%s')`, b.size))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceKafkaBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SourceKafkaBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = '%s');`, b.qualifiedName(), newSize)
}

func (b *SourceKafkaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s;`, b.qualifiedName())
}

func (b *SourceKafkaBuilder) ReadId() string {
	return readSourceId(b.sourceName, b.schemaName, b.databaseName)
}

func sourceKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourceKafkaBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		builder.KafkaConnection(v.(string))
	}

	if v, ok := d.GetOk("topic"); ok {
		builder.Topic(v.(string))
	}

	if v, ok := d.GetOk("include_key"); ok {
		builder.IncludeKey(v.(string))
	}

	if v, ok := d.GetOk("include_headers"); ok && v.(bool) {
		builder.IncludeHeaders()
	}

	if v, ok := d.GetOk("include_partition"); ok {
		builder.IncludePartition(v.(string))
	}

	if v, ok := d.GetOk("include_offset"); ok {
		builder.IncludeOffset(v.(string))
	}

	if v, ok := d.GetOk("include_timestamp"); ok {
		builder.IncludeTimestamp(v.(string))
	}

	if v, ok := d.GetOk("format"); ok {
		builder.Format(v.(string))
	}

	if v, ok := d.GetOk("key_format"); ok {
		builder.KeyFormat(v.(string))
	}

	if v, ok := d.GetOk("envelope"); ok {
		builder.Envelope(v.(string))
	}

	if v, ok := d.GetOk("schema_registry_connection"); ok {
		builder.SchemaRegistryConnection(v.(string))
	}

	if v, ok := d.GetOk("key_strategy"); ok {
		builder.KeyStrategy(v.(string))
	}

	if v, ok := d.GetOk("value_strategy"); ok {
		builder.ValueStrategy(v.(string))
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
	return SourceRead(ctx, d, meta)
}

func sourceKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSourceKafkaBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSourceKafkaBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return SourceRead(ctx, d, meta)
}

func sourceKafkaDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newSourceKafkaBuilder(sourceName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "source"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
