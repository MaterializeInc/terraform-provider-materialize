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

	// If host has changed alter the connection
	if d.HasChange("host") {
		oldHost, newHost := d.GetChange("host")
		b := materialize.NewConnection(metaDb, o)
		if err := b.Alter("HOST", newHost, false, validate); err != nil {
			d.Set("host", oldHost)
			return diag.FromErr(err)
		}
	}

	// If port has changed alter the connection
	if d.HasChange("PORT") {
		oldPort, newPort := d.GetChange("port")
		b := materialize.NewConnection(metaDb, o)
		if err := b.Alter("port", newPort, false, validate); err != nil {
			d.Set("port", oldPort)
			return diag.FromErr(err)
		}
	}

	// If user has changed alter the connection
	if d.HasChange("user") {
		oldUser, newUser := d.GetChange("user")
		user := materialize.GetValueSecretStruct(newUser)
		b := materialize.NewConnection(metaDb, o)
		if err := b.Alter("USER", user, false, validate); err != nil {
			d.Set("user", oldUser)
			return diag.FromErr(err)
		}
	}

	// If password has changed alter the connection
	if d.HasChange("password") {
		oldPassword, newPassword := d.GetChange("password")
		password := materialize.GetIdentifierSchemaStruct(newPassword)
		b := materialize.NewConnection(metaDb, o)
		if err := b.Alter("PASSWORD", password, true, validate); err != nil {
			d.Set("password", oldPassword)
			return diag.FromErr(err)
		}
	}

	// If database has changed alter the connection
	if d.HasChange("database") {
		oldDatabase, newDatabase := d.GetChange("database")
		b := materialize.NewConnection(metaDb, o)
		if err := b.Alter("DATABASE", newDatabase, false, validate); err != nil {
			d.Set("database", oldDatabase)
			return diag.FromErr(err)
		}
	}

	// If SSH tunnel has changed alter the connection
	if d.HasChange("ssh_tunnel") {
		oldTunnel, newTunnel := d.GetChange("ssh_tunnel")
		b := materialize.NewConnection(metaDb, o)
		if newTunnel == nil || len(newTunnel.([]interface{})) == 0 {
			if err := b.AlterDrop("SSH TUNNEL", validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
				return diag.FromErr(err)
			}
		} else {
			tunnel := materialize.GetIdentifierSchemaStruct(newTunnel)
			if err := b.Alter("SSH TUNNEL", tunnel, false, validate); err != nil {
				d.Set("ssh_tunnel", oldTunnel)
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
