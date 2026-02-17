package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourceKafkaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("source"),
	"size":               ObjectSizeSchema("source"),
	"kafka_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "kafka_connection",
		Description: "The Kafka connection to use in the source.",
		Required:    true,
		ForceNew:    true,
	}),
	"topic": {
		Description: "The Kafka topic you want to subscribe to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"include_key": {
		Description: "(Deprecated) Include a column containing the Kafka message key. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_key` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_key_alias": {
		Description: "(Deprecated) Provide an alias for the key column. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_key_alias` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_headers": {
		Description: "(Deprecated) Include message headers. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_headers` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"include_headers_alias": {
		Description: "(Deprecated) Provide an alias for the headers column. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_headers_alias` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_partition": {
		Description: "(Deprecated) Include a partition column containing the Kafka message partition. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_partition` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_partition_alias": {
		Description: "(Deprecated) Provide an alias for the partition column. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_partition_alias` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_offset": {
		Description: "(Deprecated) Include an offset column containing the Kafka message offset. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_offset` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_offset_alias": {
		Description: "(Deprecated) Provide an alias for the offset column. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_offset_alias` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_timestamp": {
		Description: "(Deprecated) Include a timestamp column containing the Kafka message timestamp. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_timestamp` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_timestamp_alias": {
		Description: "(Deprecated) Provide an alias for the timestamp column. Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `include_timestamp_alias` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"format": func() *schema.Schema {
		s := FormatSpecSchema("format", "(Deprecated) How to decode raw bytes from different formats into data structures Materialize can understand at runtime. Use `materialize_source_table_kafka` resources instead.", false)
		s.Deprecated = "The `format` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead."
		return s
	}(),
	"key_format": func() *schema.Schema {
		s := FormatSpecSchema("key_format", "(Deprecated) Set the key format explicitly. Use `materialize_source_table_kafka` resources instead.", false)
		s.Deprecated = "The `key_format` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead."
		return s
	}(),
	"value_format": func() *schema.Schema {
		s := FormatSpecSchema("value_format", "(Deprecated) Set the value format explicitly. Use `materialize_source_table_kafka` resources instead.", false)
		s.Deprecated = "The `value_format` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead."
		return s
	}(),
	"envelope": {
		Description: "(Deprecated) How Materialize should interpret records (e.g. append-only, upsert). Use `materialize_source_table_kafka` resources instead.",
		Deprecated:  "The `envelope` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_kafka` resources instead.",
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
					ConflictsWith: []string{"envelope.0.upsert", "envelope.0.none", "envelope.0.upsert_options"},
				},
				"none": {
					Description:   "Use an append-only envelope. This means that records will only be appended and cannot be updated or deleted.",
					Type:          schema.TypeBool,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"envelope.0.upsert", "envelope.0.debezium", "envelope.0.upsert_options"},
				},
				"upsert_options": {
					Description: "Options for the upsert envelope.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					ForceNew:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"value_decoding_errors": {
								Description: "Specify how to handle value decoding errors in the upsert envelope.",
								Type:        schema.TypeList,
								MaxItems:    1,
								Optional:    true,
								ForceNew:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"inline": {
											Description: "Configuration for inline value decoding errors.",
											Type:        schema.TypeList,
											MaxItems:    1,
											Optional:    true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"enabled": {
														Description: "Enable inline value decoding errors.",
														Type:        schema.TypeBool,
														Optional:    true,
														Default:     false,
													},
													"alias": {
														Description: "Specify an alias for the value decoding errors column, to use an alternative name for the error column. If not specified, the column name will be `error`.",
														Type:        schema.TypeString,
														Optional:    true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Optional: true,
		ForceNew: true,
	},
	"start_offset": {
		Description:   "Read partitions from the specified offset.",
		Type:          schema.TypeList,
		Elem:          &schema.Schema{Type: schema.TypeInt},
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"start_timestamp"},
	},
	"start_timestamp": {
		Description:   "Use the specified value to set `START OFFSET` based on the Kafka timestamp.",
		Type:          schema.TypeInt,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"start_offset"},
	},
	"expose_progress": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "expose_progress",
		Description: "The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`.",
		Required:    false,
		ForceNew:    true,
	}),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka source describes a Kafka cluster you want Materialize to read data from.",

		CreateContext: sourceKafkaCreate,
		ReadContext:   sourceKafkaRead,
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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceKafkaBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("kafka_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.KafkaConnection(conn)
	}

	if v, ok := d.GetOk("topic"); ok {
		b.Topic(v.(string))
	}

	if v, ok := d.GetOk("include_key"); ok && v.(bool) {
		if alias, ok := d.GetOk("include_key_alias"); ok {
			b.IncludeKeyAlias(alias.(string))
		} else {
			b.IncludeKey()
		}
	}

	if v, ok := d.GetOk("include_partition"); ok && v.(bool) {
		if alias, ok := d.GetOk("include_partition_alias"); ok {
			b.IncludePartitionAlias(alias.(string))
		} else {
			b.IncludePartition()
		}
	}

	if v, ok := d.GetOk("include_offset"); ok && v.(bool) {
		if alias, ok := d.GetOk("include_offset_alias"); ok {
			b.IncludeOffsetAlias(alias.(string))
		} else {
			b.IncludeOffset()
		}
	}

	if v, ok := d.GetOk("include_timestamp"); ok && v.(bool) {
		if alias, ok := d.GetOk("include_timestamp_alias"); ok {
			b.IncludeTimestampAlias(alias.(string))
		} else {
			b.IncludeTimestamp()
		}
	}

	if v, ok := d.GetOk("include_headers"); ok && v.(bool) {
		if alias, ok := d.GetOk("include_headers_alias"); ok {
			b.IncludeHeadersAlias(alias.(string))
		} else {
			b.IncludeHeaders()
		}
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
		envelope := materialize.GetSourceKafkaEnvelopeStruct(v)
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
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
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
	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceRead(ctx, d, meta)
}

func sourceKafkaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanSource(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.SourceName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Source{SourceName: s.SourceName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	if s.ConnectionName.Valid && s.ConnectionSchemaName.Valid && s.ConnectionDatabaseName.Valid {
		kafkaConnection := []interface{}{
			map[string]interface{}{
				"name":          s.ConnectionName.String,
				"schema_name":   s.ConnectionSchemaName.String,
				"database_name": s.ConnectionDatabaseName.String,
			},
		}
		if err := d.Set("kafka_connection", kafkaConnection); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
