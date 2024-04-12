package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func connectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanConnection(metaDb, utils.ExtractId(i))
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

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: s.ConnectionName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if d.HasChange("host") {
		oldHost, newHost := d.GetChange("host")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"HOST": newHost,
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("host", oldHost)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("port") {
		oldPort, newPort := d.GetChange("port")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"PORT": newPort,
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("port", oldPort)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("user") {
		oldUser, newUser := d.GetChange("user")
		user := materialize.GetValueSecretStruct(newUser)
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"USER": user,
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("user", oldUser)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		oldPassword, newPassword := d.GetChange("password")
		password := materialize.GetIdentifierSchemaStruct(newPassword)
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"PASSWORD": password,
		}
		if err := b.Alter(options, nil, true, validate); err != nil {
			d.Set("password", oldPassword)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("database") {
		oldDatabase, newDatabase := d.GetChange("database")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"DATABASE": newDatabase,
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("database", oldDatabase)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ssh_tunnel") {
		oldTunnel, newTunnel := d.GetChange("ssh_tunnel")
		b := materialize.NewConnection(metaDb, o)
		if newTunnel == nil || len(newTunnel.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"SSH TUNNEL"}, validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
				return diag.FromErr(err)
			}
		} else {
			tunnel := materialize.GetIdentifierSchemaStruct(newTunnel)
			options := map[string]interface{}{
				"SSH TUNNEL": tunnel,
			}
			if err := b.Alter(options, nil, false, validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
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

	if d.HasChange("ssl_mode") {
		oldSslMode, newSslMode := d.GetChange("ssl_mode")
		b := materialize.NewConnection(metaDb, o)
		if newSslMode == nil || newSslMode == "" {
			if err := b.AlterDrop([]string{"SSL MODE"}, validate); err != nil {
				d.Set("ssl_mode", oldSslMode)
				return diag.FromErr(err)
			}
		} else {
			options := map[string]interface{}{
				"SSL MODE": newSslMode,
			}
			if err := b.Alter(options, nil, false, validate); err != nil {
				d.Set("ssl_mode", oldSslMode)
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("aws_privatelink") {
		oldAwsPrivatelink, newAwsPrivatelink := d.GetChange("aws_privatelink")
		b := materialize.NewConnection(metaDb, o)
		if newAwsPrivatelink == nil || len(newAwsPrivatelink.([]interface{})) == 0 {
			if err := b.AlterDrop([]string{"AWS PRIVATELINK"}, validate); err != nil {
				d.Set("aws_privatelink", oldAwsPrivatelink)
				return diag.FromErr(err)
			}
		} else {
			awsPrivatelink := materialize.GetIdentifierSchemaStruct(newAwsPrivatelink)
			options := map[string]interface{}{
				"AWS PRIVATELINK": awsPrivatelink,
			}
			if err := b.Alter(options, nil, false, validate); err != nil {
				d.Set("aws_privatelink", oldAwsPrivatelink)
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

	return connectionRead(ctx, d, meta)
}

func connectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnection(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
