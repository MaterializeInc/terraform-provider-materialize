package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var sourceWebhookSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name": {
		Description: "The cluster to maintain this source.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"size": {
		Description: "The size of the cluster maintaining this source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"body_format": {
		Description: "The body format of the webhook.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		ValidateFunc: validation.StringInSlice([]string{
			"TEXT",
			"JSON",
			"BYTES",
		}, true),
	},
	"include_header": {
		Description: "Map a header value from a request into a column.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"header": {
					Description: "The name for the header.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"alias": {
					Description: "The alias for the header.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"bytes": {
					Description: "Change type to `bytea`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			},
		},
		ForceNew: true,
	},
	"include_headers": {
		Description: "Include headers in the webhook.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"all": {
					Description:   "Include all headers.",
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{"include_headers.0.only", "include_headers.0.not"},
					AtLeastOneOf:  []string{"include_headers.0.all", "include_headers.0.only", "include_headers.0.not"},
				},
				"only": {
					Description:   "Headers that should be included.",
					Type:          schema.TypeList,
					Elem:          &schema.Schema{Type: schema.TypeString},
					Optional:      true,
					ConflictsWith: []string{"include_headers.0.all"},
					AtLeastOneOf:  []string{"include_headers.0.all", "include_headers.0.only", "include_headers.0.not"},
				},
				"not": {
					Description:   "Headers that should be excluded.",
					Type:          schema.TypeList,
					Elem:          &schema.Schema{Type: schema.TypeString},
					Optional:      true,
					ConflictsWith: []string{"include_headers.0.all"},
					AtLeastOneOf:  []string{"include_headers.0.all", "include_headers.0.only", "include_headers.0.not"},
				},
			},
		},
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		ForceNew: true,
	},
	"check_options": {
		Description: "The check options for the webhook.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Description: "The field for the check options.",
					Type:        schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"body": {
								Description: "The body for the check options.",
								Type:        schema.TypeBool,
								Optional:    true,
							},
							"headers": {
								Description: "The headers for the check options.",
								Type:        schema.TypeBool,
								Optional:    true,
							},
							"secret": IdentifierSchema(IdentifierSchemaParams{
								Elem:        "secret",
								Description: "The secret for the check options.",
								Required:    false,
								ForceNew:    true,
							}),
						},
					},
					MinItems: 1,
					MaxItems: 1,
					Required: true,
				},
				"alias": {
					Description: "The alias for the check options.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"bytes": {
					Description: "Change type to `bytea`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			},
		},
		ForceNew: true,
	},
	"check_expression": {
		Description: "The check expression for the webhook.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "A webhook source describes a webhook you want Materialize to read data from. " +
			"This resource is deprecated and will be removed in a future release. " +
			"Please use materialize_source_table_webhook instead.",

		DeprecationMessage: "This resource is deprecated and will be removed in a future release. " +
			"Please use materialize_source_table_webhook instead.",

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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceWebhookBuilder(metaDb, o)

	b.ClusterName(clusterName).
		BodyFormat(bodyFormat).
		CheckExpression(d.Get("check_expression").(string))

	if v, ok := d.GetOk("include_header"); ok {
		var headers []materialize.HeaderStruct
		for _, header := range v.([]interface{}) {
			h := header.(map[string]interface{})
			headers = append(headers, materialize.HeaderStruct{
				Header: h["header"].(string),
				Alias:  h["alias"].(string),
				Bytes:  h["bytes"].(bool),
			})
		}
		b.IncludeHeader(headers)
	}

	if v, ok := d.GetOk("include_headers"); ok {
		var i materialize.IncludeHeadersStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["all"]; ok {
			i.All = v.(bool)
		}

		if v, ok := u["only"]; ok {
			o, err := materialize.GetSliceValueString("only", v.([]interface{}))
			if err != nil {
				return diag.FromErr(err)
			}
			i.Only = o
		}

		if v, ok := u["not"]; ok {
			n, err := materialize.GetSliceValueString("not", v.([]interface{}))
			if err != nil {
				return diag.FromErr(err)
			}
			i.Not = n
		}
		b.IncludeHeaders(i)
	}

	if v, ok := d.GetOk("check_options"); ok {
		var options []materialize.CheckOptionsStruct
		for _, option := range v.([]interface{}) {
			t := option.(map[string]interface{})
			fieldMap := t["field"].([]interface{})[0].(map[string]interface{})

			var secret = materialize.IdentifierSchemaStruct{}
			if secretMap, ok := fieldMap["secret"].([]interface{}); ok && len(secretMap) > 0 && secretMap[0] != nil {
				secret = materialize.GetIdentifierSchemaStruct(secretMap)
			}

			field := materialize.FieldStruct{
				Body:    fieldMap["body"].(bool),
				Headers: fieldMap["headers"].(bool),
				Secret:  secret,
			}

			options = append(options, materialize.CheckOptionsStruct{
				Field: field,
				Alias: t["alias"].(string),
				Bytes: t["bytes"].(bool),
			})
		}
		b.CheckOptions(options)
	}
	// Create resource
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

	// Set id
	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceRead(ctx, d, meta)
}
