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

var connectionAwsSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"endpoint": {
		Description: "Override the default AWS endpoint URL.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
	},
	"aws_region": {
		Description: "The AWS region to connect to.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
	},
	"access_key_id": ValueSecretSchema("access_key_id", "The access key ID to connect with.", false, false),
	"secret_access_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "secret_access_key",
		Description: "The secret access key corresponding to the specified access key ID.",
		Required:    false,
		ForceNew:    false,
	}),
	"session_token": ValueSecretSchema("session_token", "The session token corresponding to the specified access key ID.", false, false),
	"assume_role_arn": {
		Description: "The Amazon Resource Name (ARN) of the IAM role to assume.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
	},
	"assume_role_session_name": {
		Description:  "The session name to use when assuming the role.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     false,
		RequiredWith: []string{"assume_role_arn"},
	},
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionAws() *schema.Resource {
	return &schema.Resource{
		Description: "An Amazon Web Services (AWS) connection provides Materialize with access to an Identity and Access Management (IAM) user or role in your AWS account.",

		CreateContext: connectionAwsCreate,
		ReadContext:   connectionAwsRead,
		UpdateContext: connectionAwsUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionAwsSchema,
	}
}

func connectionAwsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanConnectionAws(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.ConnectionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("endpoint", s.Endpoint.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("aws_region", s.AwsRegion.String); err != nil {
		return diag.FromErr(err)
	}

	// Note: Cannot set nested attributes with this SDK
	// https://github.com/hashicorp/terraform-plugin-sdk/issues/459
	// if err := d.Set("access_key_id.0.text", s.AccessKeyId.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	// if err := d.Set("access_key_id.secret", s.AccessKeyIdSecretId.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	// if err := d.Set("secret_access_key", s.AccessKeyIdSecretId.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	// if err := d.Set("session_token.text", s.SessionToken.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	// if err := d.Set("session_token.secret", s.SessionTokenSecretId.String); err != nil {
	// 	return diag.FromErr(err)
	// }

	if err := d.Set("assume_role_arn", s.AssumeRoleArn.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("assume_role_session_name", s.AssumeRoleSessionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: s.ConnectionName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionAwsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionAwsBuilder(metaDb, o)

	if v, ok := d.GetOk("endpoint"); ok {
		b.Endpoint(v.(string))
	}

	if v, ok := d.GetOk("aws_region"); ok {
		b.AwsRegion(v.(string))
	}

	if v, ok := d.GetOk("access_key_id"); ok {
		a := materialize.GetValueSecretStruct(v)
		b.AccessKeyId(a)
	}

	if v, ok := d.GetOk("secret_access_key"); ok {
		s := materialize.GetIdentifierSchemaStruct(v)
		b.SecretAccessKey(s)
	}

	if v, ok := d.GetOk("session_token"); ok {
		s := materialize.GetValueSecretStruct(v)
		b.SessionToken(s)
	}

	if v, ok := d.GetOk("assume_role_arn"); ok {
		b.AssumeRoleArn(v.(string))
	}

	if v, ok := d.GetOk("assume_role_session_name"); ok {
		b.AssumeRoleSessionName(v.(string))
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
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
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionAwsRead(ctx, d, meta)
}

func connectionAwsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	validate := d.Get("validate").(bool)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewConnectionAwsBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("endpoint") {
		oldEndpoint, newEndpoint := d.GetChange("endpoint")
		b := materialize.NewConnection(metaDb, o)
		if newEndpoint == nil || newEndpoint == "" {
			if err := b.AlterDrop([]string{"ENDPOINT"}, validate); err != nil {
				d.Set("endpoint", oldEndpoint)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"ENDPOINT": newEndpoint.(string),
			}
			if err := b.Alter(options, false, validate); err != nil {
				d.Set("host", oldEndpoint)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("aws_region") {
		oldRegion, newRegion := d.GetChange("aws_region")
		b := materialize.NewConnection(metaDb, o)
		if newRegion == nil || newRegion == "" {
			if err := b.AlterDrop([]string{"REGION"}, validate); err != nil {
				d.Set("aws_region", oldRegion)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"REGION": newRegion.(string),
			}
			if err := b.Alter(options, false, validate); err != nil {
				d.Set("aws_region", oldRegion)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("access_key_id") {
		oldAccessKeyId, newAccessKeyId := d.GetChange("access_key_id")
		b := materialize.NewConnectionAwsBuilder(metaDb, o)
		if newAccessKeyId == nil || newAccessKeyId == "" || len(newAccessKeyId.([]interface{})) == 0 {
			// TODO: Can't drop access key id without secret session token
			if err := b.AlterDrop([]string{"ACCESS KEY ID"}, validate); err != nil {
				d.Set("access_key_id", oldAccessKeyId)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"ACCESS KEY ID": materialize.GetValueSecretStruct(newAccessKeyId),
			}
			if err := b.Alter(options, false, validate); err != nil {
				d.Set("access_key_id", oldAccessKeyId)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("secret_access_key") {
		oldSecretAccessKey, newSecretAccessKey := d.GetChange("secret_access_key")
		b := materialize.NewConnectionAwsBuilder(metaDb, o)
		if newSecretAccessKey == nil || newSecretAccessKey == "" || len(newSecretAccessKey.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"SECRET ACCESS KEY"}, validate); err != nil {
				d.Set("secret_access_key", oldSecretAccessKey)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"SECRET ACCESS KEY": materialize.GetIdentifierSchemaStruct(newSecretAccessKey),
			}
			if err := b.Alter(options, true, validate); err != nil {
				d.Set("secret_access_key", oldSecretAccessKey)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("session_token") {
		oldSessionToken, newSessionToken := d.GetChange("session_token")
		b := materialize.NewConnectionAwsBuilder(metaDb, o)
		if newSessionToken == nil || newSessionToken == "" || len(newSessionToken.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"SESSION TOKEN"}, validate); err != nil {
				d.Set("session_token", oldSessionToken)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"SESSION TOKEN": materialize.GetValueSecretStruct(newSessionToken),
			}
			if err := b.Alter(options, false, validate); err != nil {
				d.Set("session_token", oldSessionToken)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("assume_role_arn") || d.HasChange("assume_role_session_name") {
		oldAssumeRoleArn, newAssumeRoleArn := d.GetChange("assume_role_arn")
		oldAssumeRoleSessionName, newAssumeRoleSessionName := d.GetChange("assume_role_session_name")
		b := materialize.NewConnectionAwsBuilder(metaDb, o)
		options := make(map[string]interface{})

		if newAssumeRoleArn != nil && newAssumeRoleArn != "" {
			options["ASSUME ROLE ARN"] = newAssumeRoleArn.(string)
		} else if d.HasChange("assume_role_arn") {
			options["ASSUME ROLE ARN"] = ""
		}

		if newAssumeRoleSessionName != nil && newAssumeRoleSessionName != "" {
			options["ASSUME ROLE SESSION NAME"] = newAssumeRoleSessionName.(string)
		} else if d.HasChange("assume_role_session_name") {
			options["ASSUME ROLE SESSION NAME"] = ""
		}

		// Perform the alteration only if there are options to update
		if len(options) > 0 {
			if err := b.Alter(options, false, validate); err != nil {
				d.Set("assume_role_arn", oldAssumeRoleArn)
				d.Set("assume_role_session_name", oldAssumeRoleSessionName)
				return diag.FromErr(err)
			}
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

	return connectionAwsRead(ctx, d, meta)
}
