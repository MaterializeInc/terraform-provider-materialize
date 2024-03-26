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
		ForceNew:    true,
	},
	"aws_region": {
		Description: "The AWS region to connect to.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"access_key_id": ValueSecretSchema("access_key_id", "The access key ID to connect with.", false, true),
	"secret_access_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "secret_access_key",
		Description: "The secret access key corresponding to the specified access key ID.",
		Required:    false,
		ForceNew:    true,
	}),
	"session_token": ValueSecretSchema("session_token", "The session token corresponding to the specified access key ID.", false, true),
	"assume_role_arn": {
		Description: "The Amazon Resource Name (ARN) of the IAM role to assume.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"assume_role_session_name": {
		Description:  "The session name to use when assuming the role.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
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
