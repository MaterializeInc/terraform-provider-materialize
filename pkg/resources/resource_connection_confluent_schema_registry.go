package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var connectionConfluentSchemaRegistrySchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"url": {
		Description: "The URL of the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Confluent Schema Registry.", false, false),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Confluent Schema Registry.", false, false),
	"ssl_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssl_key",
		Description: "The client key for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    false,
	}),
	"password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "password",
		Description: "The password for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    false,
	}),
	"username": ValueSecretSchema("username", "The username for the Confluent Schema Registry.", false, false),
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The SSH tunnel configuration for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the Confluent Schema Registry.",
		Required:    false,
		ForceNew:    true,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionConfluentSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "A Confluent Schema Registry connection establishes a link to a Confluent Schema Registry server.",

		CreateContext: connectionConfluentSchemaRegistryCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionConfluentSchemaRegistryUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionConfluentSchemaRegistrySchema,
	}
}

func connectionConfluentSchemaRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionConfluentSchemaRegistryBuilder(metaDb, o)

	if v, ok := d.GetOk("url"); ok {
		b.ConfluentSchemaRegistryUrl(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistrySSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistrySSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		key := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistrySSLKey(key)
	}

	if v, ok := d.GetOk("username"); ok {
		user := materialize.GetValueSecretStruct(v)
		b.ConfluentSchemaRegistryUsername(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistryPassword(pass)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistrySSHTunnel(conn)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.ConfluentSchemaRegistryAWSPrivateLink(conn)
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

	return connectionRead(ctx, d, meta)
}

func connectionConfluentSchemaRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		b := materialize.NewConnection(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("url") {
		oldUrl, newUrl := d.GetChange("url")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"URL": newUrl,
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("url", oldUrl)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("username") {
		oldUser, newUser := d.GetChange("username")
		b := materialize.NewConnection(metaDb, o)
		if newUser == nil || len(newUser.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"USER"}, validate); err != nil {
				d.Set("username", oldUser)
				return diag.FromErr(err)
			}
		} else {
			user := materialize.GetValueSecretStruct(newUser)
			options := map[string]interface{}{
				"USER": user,
			}
			if err := b.Alter(options, nil, false, validate); err != nil {
				d.Set("username", oldUser)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("password") {
		oldPassword, newPassword := d.GetChange("password")
		b := materialize.NewConnection(metaDb, o)
		if newPassword == nil || len(newPassword.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"PASSWORD"}, validate); err != nil {
				d.Set("password", oldPassword)
				return diag.FromErr(err)
			}
		} else {
			password := materialize.GetIdentifierSchemaStruct(newPassword)
			options := map[string]interface{}{
				"PASSWORD": password,
			}
			if err := b.Alter(options, nil, true, validate); err != nil {
				d.Set("password", oldPassword)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("ssl_certificate_authority") {
		oldSslCa, newSslCa := d.GetChange("ssl_certificate_authority")
		b := materialize.NewConnection(metaDb, o)
		if newSslCa == nil || len(newSslCa.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"SSL CERTIFICATE AUTHORITY"}, validate); err != nil {
				d.Set("ssl_certificate_authority", oldSslCa)
				return diag.FromErr(err)
			}
		} else {
			sslCa := materialize.GetValueSecretStruct(newSslCa)
			options := map[string]interface{}{
				"SSL CERTIFICATE AUTHORITY": sslCa,
			}
			if err := b.Alter(options, nil, true, validate); err != nil {
				d.Set("ssl_certificate_authority", oldSslCa)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("ssl_certificate") || d.HasChange("ssl_key") {
		oldSslCert, newSslCert := d.GetChange("ssl_certificate")
		oldSslKey, newSslKey := d.GetChange("ssl_key")
		b := materialize.NewConnection(metaDb, o)

		if (newSslCert == nil || len(newSslCert.([]interface{})) == 0) || (newSslKey == nil || len(newSslKey.([]interface{})) == 0) {
			// Drop both SSL CERTIFICATE and SSL KEY in a single statement
			if err := b.AlterDrop([]string{"SSL CERTIFICATE", "SSL KEY"}, validate); err != nil {
				d.Set("ssl_certificate", oldSslCert)
				d.Set("ssl_key", oldSslKey)
				return diag.FromErr(err)
			}
		} else {
			options := make(map[string]interface{})
			if newSslCert != nil && len(newSslCert.([]interface{})) > 0 {
				sslCert := materialize.GetValueSecretStruct(newSslCert)
				options["SSL CERTIFICATE"] = sslCert
			}
			if newSslKey != nil && len(newSslKey.([]interface{})) > 0 {
				sslKey := materialize.GetIdentifierSchemaStruct(newSslKey)
				options["SSL KEY"] = sslKey
			}
			if len(options) > 0 {
				if err := b.Alter(options, nil, true, validate); err != nil {
					d.Set("ssl_certificate", oldSslCert)
					d.Set("ssl_key", oldSslKey)
					return diag.FromErr(err)
				}
			}
		}
	}

	// TODO: Uncomment when SSH TUNNEL and AWS PRIVATELINK are supported in the Materialize
	// if d.HasChange("ssh_tunnel") {
	// 	oldTunnel, newTunnel := d.GetChange("ssh_tunnel")
	// 	b := materialize.NewConnection(metaDb, o)
	// 	if newTunnel == nil || len(newTunnel.([]interface{})) == 0 {
	// 		if err := b.AlterDrop([]string{"SSH TUNNEL"}, validate); err != nil {
	// 			d.Set("ssh_tunnel", oldTunnel)
	// 			return diag.FromErr(err)
	// 		}
	// 	} else {
	// 		tunnel := materialize.GetIdentifierSchemaStruct(newTunnel)
	// 		options := map[string]interface{}{
	// 			"SSH TUNNEL": tunnel,
	// 		}
	// 		if err := b.Alter(options, nil, false, validate); err != nil {
	// 			d.Set("ssh_tunnel", oldTunnel)
	// 			return diag.FromErr(err)
	// 		}
	// 	}
	// }

	// if d.HasChange("aws_privatelink") {
	// 	oldAwsPrivatelink, newAwsPrivatelink := d.GetChange("aws_privatelink")
	// 	b := materialize.NewConnection(metaDb, o)
	// 	if newAwsPrivatelink == nil || len(newAwsPrivatelink.([]interface{})) == 0 {
	// 		if err := b.AlterDrop([]string{"AWS PRIVATELINK"}, validate); err != nil {
	// 			d.Set("aws_privatelink", oldAwsPrivatelink)
	// 			return diag.FromErr(err)
	// 		}
	// 	} else {
	// 		awsPrivatelink := materialize.GetIdentifierSchemaStruct(newAwsPrivatelink)
	// 		options := map[string]interface{}{
	// 			"AWS PRIVATELINK": awsPrivatelink,
	// 		}
	// 		if err := b.Alter(options, nil, false, validate); err != nil {
	// 			d.Set("aws_privatelink", oldAwsPrivatelink)
	// 			return diag.FromErr(err)
	// 		}
	// 	}
	// }

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

	return connectionRead(ctx, d, meta)
}
