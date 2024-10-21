package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourceTableKafkaSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source table", true, false),
	"schema_name":        SchemaNameSchema("source table", false),
	"database_name":      DatabaseNameSchema("source table", false),
	"qualified_sql_name": QualifiedNameSchema("source table"),
	"comment":            CommentSchema(false),
	"source": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "source",
		Description: "The source this table is created from.",
		Required:    true,
		ForceNew:    true,
	}),
	"topic": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Computed:    true,
		Description: "The name of the Kafka topic in the Kafka cluster.",
	},
	"include_key": {
		Description: "Include a column containing the Kafka message key.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_key_alias": {
		Description: "Provide an alias for the key column.",
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
	"include_headers_alias": {
		Description: "Provide an alias for the headers column.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_partition": {
		Description: "Include a partition column containing the Kafka message partition",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_partition_alias": {
		Description: "Provide an alias for the partition column.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_offset": {
		Description: "Include an offset column containing the Kafka message offset.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_offset_alias": {
		Description: "Provide an alias for the offset column.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"include_timestamp": {
		Description: "Include a timestamp column containing the Kafka message timestamp.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"include_timestamp_alias": {
		Description: "Provide an alias for the timestamp column.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
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
	"expose_progress": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "expose_progress",
		Description: "The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`.",
		Required:    false,
		ForceNew:    true,
	}),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceTableKafka() *schema.Resource {
	return &schema.Resource{
		Description: "A Kafka source describes a Kafka cluster you want Materialize to read data from.",

		CreateContext: sourceTableKafkaCreate,
		ReadContext:   sourceTableKafkaRead,
		UpdateContext: sourceTableKafkaUpdate,
		DeleteContext: sourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceTableKafkaSchema,
	}
}

func sourceTableKafkaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceTableKafkaBuilder(metaDb, o)

	source := materialize.GetIdentifierSchemaStruct(d.Get("source"))
	b.Source(source)

	b.UpstreamName(d.Get("topic").(string))

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

	if v, ok := d.GetOk("expose_progress"); ok {
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.SourceTableKafkaId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceTableKafkaRead(ctx, d, meta)
}

func sourceTableKafkaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	t, err := materialize.ScanSourceTableKafka(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", t.TableName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", t.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", t.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	source := []interface{}{
		map[string]interface{}{
			"name":          t.SourceName.String,
			"schema_name":   t.SourceSchemaName.String,
			"database_name": t.SourceDatabaseName.String,
		},
	}
	if err := d.Set("source", source); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("topic", t.UpstreamName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", t.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", t.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sourceTableKafkaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "TABLE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSourceTableKafkaBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return sourceTableKafkaRead(ctx, d, meta)
}
