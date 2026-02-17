package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var sourceTableWebhookSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("table", true, false),
	"schema_name":        SchemaNameSchema("table", false),
	"database_name":      DatabaseNameSchema("table", false),
	"qualified_sql_name": QualifiedNameSchema("table"),
	"comment":            CommentSchema(false),
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

func SourceTableWebhook() *schema.Resource {
	return &schema.Resource{
		Description: "A webhook source table allows reading data directly from webhooks.",

		CreateContext: sourceTableWebhookCreate,
		ReadContext:   sourceTableWebhookRead,
		UpdateContext: sourceTableWebhookUpdate,
		DeleteContext: sourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceTableWebhookSchema,
	}
}

func sourceTableWebhookCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Table, Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceTableWebhookBuilder(metaDb, o)

	b.BodyFormat(d.Get("body_format").(string))

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

	if v, ok := d.GetOk("check_expression"); ok {
		b.CheckExpression(v.(string))
	}

	// Create resource
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

	// Set ID
	i, err := materialize.SourceTableWebhookId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceTableWebhookRead(ctx, d, meta)
}

func sourceTableWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	t, err := materialize.ScanSourceTableWebhook(metaDb, utils.ExtractId(i))
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

	if err := d.Set("ownership_role", t.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", t.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sourceTableWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.Table, Name: tableName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.Table, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSourceTableBuilder(metaDb, o)
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

	return sourceTableWebhookRead(ctx, d, meta)
}
