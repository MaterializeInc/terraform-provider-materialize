package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourceTableLoadGenSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("table", true, false),
	"schema_name":        SchemaNameSchema("table", false),
	"database_name":      DatabaseNameSchema("table", false),
	"qualified_sql_name": QualifiedNameSchema("table"),
	"source": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "source",
		Description: "The source this table is created from.",
		Required:    true,
		ForceNew:    true,
	}),
	"upstream_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The name of the table in the upstream database.",
	},
	"upstream_schema_name": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "The schema of the table in the upstream database.",
	},
	"comment":        CommentSchema(false),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceTableLoadGen() *schema.Resource {
	return &schema.Resource{
		CreateContext: sourceTableLoadGenCreate,
		ReadContext:   sourceTableRead,
		UpdateContext: sourceTableUpdate,
		DeleteContext: sourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceTableLoadGenSchema,
	}
}

func sourceTableLoadGenCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "TABLE", Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceTableLoadGenBuilder(metaDb, o)

	source := materialize.GetIdentifierSchemaStruct(d.Get("source"))
	b.Source(source)

	b.UpstreamName(d.Get("upstream_name").(string))

	if v, ok := d.GetOk("upstream_schema_name"); ok {
		b.UpstreamSchemaName(v.(string))
	}

	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// Handle ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)
		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// Handle comments
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	i, err := materialize.SourceTableId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceTableRead(ctx, d, meta)
}
