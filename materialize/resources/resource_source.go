package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceSchema = map[string]*schema.Schema{
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
	"cluster_name": {
		Description:   "The cluster to maintain this source. If not specified, the size option must be specified.",
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"size"},
	},
	"size": {
		Description:   "The size of the source.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ValidateFunc:  validation.StringInSlice(append(sourceSizes, localSizes...), true),
		ConflictsWith: []string{"cluster_name"},
	},
	"connection_type": {
		Description:  "The source connection type.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(sourceConnectionTypes, true),
	},
	// Load Generator
	"load_generator_type": {
		Description:   "The identifier for the secret schema.",
		Type:          schema.TypeString,
		Optional:      true,
		Default:       "public",
		ValidateFunc:  validation.StringInSlice(loadGeneratorTypes, true),
		ConflictsWith: []string{"postgres_connection", "publication"},
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
	// Postgres
	"postgres_connection": {
		Description:   "The name of the PostgreSQL connection to use in the source.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"kafka_connection", "load_generator_type"},
		RequiredWith:  []string{"postgres_connection", "publication"},
	},
	"publication": {
		Description:   "The PostgreSQL publication (the replication data set containing the tables to be streamed to Materialize).",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"kafka_connection", "load_generator_type"},
		RequiredWith:  []string{"postgres_connection", "publication"},
	},
	"tables": {
		Description: "Creates subsources for specific tables in the load generator.",
		Type:        schema.TypeMap,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(replicaSizes, true),
	},
	// Broker
	"kafka_connection": {
		Description:   "The name of the Kafka connection to use in the source.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"load_generator_type", "postgres_connection"},
		RequiredWith:  []string{"kafka_connection", "topic"},
	},
	"topic": {
		Description:   "The Kafka topic you want to subscribe to.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"load_generator_type", "postgres_connection"},
		RequiredWith:  []string{"kafka_connection", "topic"},
	},
	"include_key": {
		Description: "Include a column containing the Kafka message key. If the key is encoded using a format that includes schemas the column will take its name from the schema. For unnamed formats (e.g. TEXT), the column will be named key. ",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
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
		Description: "How to decode raw bytes from different formats into data structures it can understand at runtime",
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
}

func Source() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourceCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceUpdate,
		DeleteContext: sourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceSchema,
	}
}

type SourceBuilder struct {
	sourceName               string
	schemaName               string
	databaseName             string
	clusterName              string
	size                     string
	connectionType           string
	loadGeneratorType        string
	tickInterval             string
	scaleFactor              float64
	postgresConnection       string
	publication              string
	tables                   map[string]string
	kafkaConnection          string
	topic                    string
	includeKey               string
	includePartition         string
	includeOffset            string
	includeTimestamp         string
	format                   string
	envelope                 string
	schemaRegistryConnection string
}

func newSourceBuilder(sourceName, schemaName, databaseName string) *SourceBuilder {
	return &SourceBuilder{
		sourceName:   sourceName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SourceBuilder) ClusterName(c string) *SourceBuilder {
	b.clusterName = c
	return b
}

func (b *SourceBuilder) Size(s string) *SourceBuilder {
	b.size = s
	return b
}

func (b *SourceBuilder) ConnectionType(c string) *SourceBuilder {
	b.connectionType = c
	return b
}

func (b *SourceBuilder) LoadGeneratorType(l string) *SourceBuilder {
	b.loadGeneratorType = l
	return b
}

func (b *SourceBuilder) Tables(t map[string]string) *SourceBuilder {
	b.tables = t
	return b
}

func (b *SourceBuilder) TickInterval(t string) *SourceBuilder {
	b.tickInterval = t
	return b
}

func (b *SourceBuilder) ScaleFactor(s float64) *SourceBuilder {
	b.scaleFactor = s
	return b
}

func (b *SourceBuilder) PostgresConnection(p string) *SourceBuilder {
	b.postgresConnection = p
	return b
}

func (b *SourceBuilder) Publication(p string) *SourceBuilder {
	b.publication = p
	return b
}

func (b *SourceBuilder) KafkaConnection(k string) *SourceBuilder {
	b.kafkaConnection = k
	return b
}

func (b *SourceBuilder) Topic(t string) *SourceBuilder {
	b.topic = t
	return b
}

func (b *SourceBuilder) IncludeKey(i string) *SourceBuilder {
	b.includeKey = i
	return b
}

func (b *SourceBuilder) IncludePartition(i string) *SourceBuilder {
	b.includePartition = i
	return b
}

func (b *SourceBuilder) IncludeOffset(i string) *SourceBuilder {
	b.includeOffset = i
	return b
}

func (b *SourceBuilder) IncludeTimestamp(i string) *SourceBuilder {
	b.includeTimestamp = i
	return b
}

func (b *SourceBuilder) Format(f string) *SourceBuilder {
	b.format = f
	return b
}

func (b *SourceBuilder) Envelope(e string) *SourceBuilder {
	b.envelope = e
	return b
}

func (b *SourceBuilder) SchemaRegistryConnection(s string) *SourceBuilder {
	b.schemaRegistryConnection = s
	return b
}

func (b *SourceBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s.%s.%s`, b.databaseName, b.schemaName, b.sourceName))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	if b.connectionType != "" {
		q.WriteString(fmt.Sprintf(` FROM %s`, b.connectionType))
	}

	// Load Generator
	if b.connectionType == "LOAD GENERATOR" {
		q.WriteString(fmt.Sprintf(` %s`, b.loadGeneratorType))

		var p []string
		if b.tickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL '%s'`, b.tickInterval)
			p = append(p, t)
		}

		if b.scaleFactor != 0 {
			s := fmt.Sprintf(`SCALE FACTOR %.2f`, b.scaleFactor)
			p = append(p, s)
		}

		if len(p) != 0 {
			p := strings.Join(p[:], ", ")
			q.WriteString(fmt.Sprintf(` (%s)`, p))
		}
	}

	// Postgres
	if b.connectionType == "POSTGRES" {
		q.WriteString(fmt.Sprintf(` CONNECTION %s (PUBLICATION '%s')`, b.postgresConnection, b.publication))

		var o []string
		if len(b.tables) > 0 {
			// Need to sort tables to ensure order for tests
			var keys []string
			for k := range b.tables {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				s := fmt.Sprintf(`%s AS %s`, k, b.tables[k])
				o = append(o, s)
			}
			o := strings.Join(o[:], ", ")
			q.WriteString(fmt.Sprintf(` FOR TABLES (%s)`, o))
		} else {
			q.WriteString(` FOR ALL TABLES`)
		}
	}

	// Broker
	if b.connectionType == "KAFKA" {
		q.WriteString(fmt.Sprintf(` CONNECTION %s (TOPIC '%s')`, b.kafkaConnection, b.topic))

		if b.format != "" {
			q.WriteString(fmt.Sprintf(` FORMAT %s`, b.format))
		}

		if b.schemaRegistryConnection != "" {
			q.WriteString(fmt.Sprintf(` USING CONFLUENT SCHEMA REGISTRY CONNECTION %s`, b.schemaRegistryConnection))
		}

		if b.envelope != "" {
			q.WriteString(fmt.Sprintf(` ENVELOPE %s`, b.envelope))
		}
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = '%s')`, b.size))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_sources.id
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.sourceName, b.schemaName, b.databaseName)
}

func (b *SourceBuilder) Rename(newName string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName, b.databaseName, b.schemaName, newName)
}

func (b *SourceBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s SET (SIZE = '%s');`, b.databaseName, b.schemaName, b.sourceName, newSize)
}

func (b *SourceBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName)
}

func readSourceParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.id = '%s';`, id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _source struct {
	name            sql.NullString `db:"name"`
	schema_name     sql.NullString `db:"schema_name"`
	database_name   sql.NullString `db:"database_name"`
	source_type     sql.NullString `db:"source_type"`
	size            sql.NullString `db:"size"`
	envelope_type   sql.NullString `db:"envelope_type"`
	connection_name sql.NullString `db:"connection_name"`
	cluster_name    sql.NullString `db:"cluster_name"`
}

func sourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSourceParams(i)

	readResource(conn, d, i, q, _source{}, "source")
	setQualifiedName(d)
	return nil
}

func sourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourceBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("connection_type"); ok {
		builder.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("load_generator_type"); ok {
		builder.LoadGeneratorType(v.(string))
	}

	if v, ok := d.GetOk("tick_interval"); ok {
		builder.TickInterval(v.(string))
	}

	if v, ok := d.GetOk("tick_interval"); ok {
		builder.TickInterval(v.(string))
	}

	if v, ok := d.GetOk("scale_factor"); ok {
		builder.ScaleFactor(v.(float64))
	}

	if v, ok := d.GetOk("publication"); ok {
		builder.Publication(v.(string))
	}

	if v, ok := d.GetOk("tables"); ok {
		builder.Tables(v.(map[string]string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		builder.KafkaConnection(v.(string))
	}

	if v, ok := d.GetOk("include_key"); ok {
		builder.IncludeKey(v.(string))
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

	if v, ok := d.GetOk("envelope"); ok {
		builder.Envelope(v.(string))
	}

	if v, ok := d.GetOk("schema_registry_connection"); ok {
		builder.SchemaRegistryConnection(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "source")
	return sourceRead(ctx, d, meta)
}

func sourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSourceBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSourceBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sourceRead(ctx, d, meta)
}

func sourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourceBuilder(sourceName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "source")
	return nil
}
