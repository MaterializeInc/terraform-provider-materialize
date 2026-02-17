package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sinkIcebergSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("sink", true, false),
	"schema_name":        SchemaNameSchema("sink", false),
	"database_name":      DatabaseNameSchema("sink", false),
	"qualified_sql_name": QualifiedNameSchema("sink"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("sink"),
	"size":               ObjectSizeSchema("sink"),
	"from": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "from",
		Description: "The name of the source, table or materialized view you want to send to the sink.",
		Required:    true,
		ForceNew:    false,
	}),
	"iceberg_catalog_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "iceberg_catalog_connection",
		Description: "The name of the Iceberg catalog connection to use.",
		Required:    true,
		ForceNew:    true,
	}),
	"namespace": {
		Description: "The Iceberg namespace (database) containing the table.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"table": {
		Description: "The name of the Iceberg table to write to. If the table doesn't exist, Materialize creates it with a schema matching the source.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"aws_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_connection",
		Description: "The AWS connection for object storage access.",
		Required:    true,
		ForceNew:    true,
	}),
	"key": {
		Description: "The columns that uniquely identify rows. Required for Iceberg sinks.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		ForceNew:    true,
	},
	"key_not_enforced": {
		Description: "Disable Materialize's validation of the key's uniqueness. Use only when you have outside knowledge that the key is unique.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"commit_interval": {
		Description: "How frequently to commit snapshots to Iceberg (e.g., '10s', '1m'). Required for Iceberg sinks.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SinkIceberg() *schema.Resource {
	return &schema.Resource{
		Description: "An Iceberg sink writes data from Materialize to an Apache Iceberg table stored in object storage.",

		CreateContext: sinkIcebergCreate,
		ReadContext:   sinkRead,
		UpdateContext: sinkUpdate,
		DeleteContext: sinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sinkIcebergSchema,
	}
}

func sinkIcebergCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sinkName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseSink, Name: sinkName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSinkIcebergBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("from"); ok {
		from := materialize.GetIdentifierSchemaStruct(v)
		b.From(from)
	}

	if v, ok := d.GetOk("iceberg_catalog_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.IcebergCatalogConnection(conn)
	}

	if v, ok := d.GetOk("namespace"); ok {
		b.Namespace(v.(string))
	}

	if v, ok := d.GetOk("table"); ok {
		b.Table(v.(string))
	}

	if v, ok := d.GetOk("aws_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.AwsConnection(conn)
	}

	if v, ok := d.GetOk("key"); ok && len(v.([]interface{})) > 0 {
		keys, err := materialize.GetSliceValueString("key", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.Key(keys)
	}

	if v, ok := d.GetOk("key_not_enforced"); ok {
		b.KeyNotEnforced(v.(bool))
	}

	if v, ok := d.GetOk("commit_interval"); ok {
		b.CommitInterval(v.(string))
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
	i, err := materialize.SinkId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sinkRead(ctx, d, meta)
}
