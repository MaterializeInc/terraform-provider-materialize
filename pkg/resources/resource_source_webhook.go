package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceWebhookSchema = map[string]*schema.Schema{
	"name":               NameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"cluster_name": {
		Description: "The cluster to maintain this source.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"size": {
		Description:  "The size of the source.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
		ValidateFunc: validation.StringInSlice(append(sourceSizes, localSizes...), true),
	},
	"body_format": {
		Description: "The body format of the webhook.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"include_headers": {
		Description: "Include headers in the webhook.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"check_options": {
		Description: "The check options for the webhook.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
	},
	"check_expression": {
		Description: "The check expression for the webhook.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"subsource":      SubsourceSchema(),
	"ownership_role": OwnershipRole(),
}

func SourceWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "A webhook source describes a webhook you want Materialize to read data from.",

		CreateContext: sourceWebhookCreate,
		ReadContext:   sourceRead,
		UpdateContext: sourceUpdate,
		DeleteContext: sourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceWebhookSchema,
	}
}

func sourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	clusterName := d.Get("cluster_name").(string)
	bodyFormat := d.Get("body_format").(string)

	o := materialize.ObjectSchemaStruct{Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceWebhookBuilder(meta.(*sqlx.DB), o)

	b.ClusterName(clusterName).
		BodyFormat(bodyFormat).
		IncludeHeaders(d.Get("include_headers").(bool)).
		CheckExpression(d.Get("check_expression").(string))

	checkOptions := d.Get("check_options").([]interface{})
	checkOptionsStrings := make([]string, len(checkOptions))
	for i, option := range checkOptions {
		checkOptionsStrings[i] = option.(string)
	}

	b.CheckOptions(checkOptionsStrings)

	// Create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// Set id
	i, err := materialize.SourceId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return sourceRead(ctx, d, meta)
}
