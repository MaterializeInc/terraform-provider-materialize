package resources

import (
	"context"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

// Droppable is an interface for builders that support Drop operation
type Droppable interface {
	Drop() error
}

// rollbackCreate drops an object that was created but could not be fully
// configured, and returns cause as the resource error. A drop that itself fails
// is surfaced as a warning so the leftover object does not go unnoticed.
func rollbackCreate(builder Droppable, o materialize.MaterializeObject, stage string, cause error) diag.Diagnostics {
	log.Printf("[DEBUG] resource failed %s, dropping object: %s", stage, o.Name)

	diags := diag.FromErr(cause)
	if err := builder.Drop(); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Could not drop %q after a failed %s", o.Name, stage),
			Detail:   fmt.Sprintf("The object may still exist in Materialize and need to be dropped manually: %s", err),
		})
	}

	return diags
}

// applyOwnership applies ownership to a newly created resource.
// If the operation fails, it drops the resource and returns an error.
// This is a common pattern across connection, source, table, view, and other resources.
func applyOwnership(d *schema.ResourceData, metaDb *sqlx.DB, o materialize.MaterializeObject, builder Droppable) diag.Diagnostics {
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			return rollbackCreate(builder, o, "ownership", err)
		}
	}

	return nil
}

// applyComment applies a comment to a newly created resource.
// If the operation fails, it drops the resource and returns an error.
// This is a common pattern across connection, source, table, view, and other resources.
func applyComment(d *schema.ResourceData, metaDb *sqlx.DB, o materialize.MaterializeObject, builder Droppable) diag.Diagnostics {
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			return rollbackCreate(builder, o, "comment", err)
		}
	}

	return nil
}

// createGrant creates a grant for a given object type.
// This is the common pattern used across all grant resources (cluster, database, schema, etc.).
func createGrant(ctx context.Context, d *schema.ResourceData, meta interface{}, objectType materialize.EntityType, objectNameField string) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	objectName := d.Get(objectNameField).(string)

	obj := materialize.MaterializeObject{
		ObjectType: objectType,
		Name:       objectName,
	}

	// Add schema and database qualifiers if they exist in the resource,
	// but only if the grant is not ON a database or schema itself
	if objectType != materialize.Database && objectType != materialize.Schema {
		if v, ok := d.GetOk("schema_name"); ok {
			obj.SchemaName = v.(string)
		}
		if v, ok := d.GetOk("database_name"); ok {
			obj.DatabaseName = v.(string)
		}
	} else if objectType == materialize.Schema {
		// Schema grants need database qualifier but not schema qualifier
		if v, ok := d.GetOk("database_name"); ok {
			obj.DatabaseName = v.(string)
		}
	}

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(metaDb, roleName, privilege, obj)

	// grant resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// set grant id
	roleId, err := materialize.RoleId(metaDb, roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	i, err := materialize.ObjectId(metaDb, obj)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(string(region), i, roleId, privilege)
	d.SetId(key)

	return grantRead(ctx, d, meta)
}

// revokeGrant revokes a grant for a given object type.
// This is the common pattern used across all grant resources (cluster, database, schema, etc.).
func revokeGrant(d *schema.ResourceData, meta interface{}, objectType materialize.EntityType, objectNameField string) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	objectName := d.Get(objectNameField).(string)

	obj := materialize.MaterializeObject{
		ObjectType: objectType,
		Name:       objectName,
	}

	// Add schema and database qualifiers if they exist in the resource,
	// but only if the grant is not ON a database or schema itself
	if objectType != materialize.Database && objectType != materialize.Schema {
		if v, ok := d.GetOk("schema_name"); ok {
			obj.SchemaName = v.(string)
		}
		if v, ok := d.GetOk("database_name"); ok {
			obj.DatabaseName = v.(string)
		}
	} else if objectType == materialize.Schema {
		// Schema grants need database qualifier but not schema qualifier
		if v, ok := d.GetOk("database_name"); ok {
			obj.DatabaseName = v.(string)
		}
	}

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(metaDb, roleName, privilege, obj)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// createDefaultPrivilegeGrant creates a default privilege grant for a given object type.
// This is the common pattern used across all default privilege grant resources.
func createDefaultPrivilegeGrant(ctx context.Context, d *schema.ResourceData, meta interface{}, objectType materialize.EntityType) diag.Diagnostics {
	granteeName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, objectType, granteeName, targetName, privilege)

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// Query ids
	gId, err := materialize.RoleId(metaDb, granteeName)
	if err != nil {
		return diag.FromErr(err)
	}

	tId, err := materialize.RoleId(metaDb, targetName)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(string(region), string(objectType), gId, tId, "", "", privilege)
	d.SetId(key)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

// revokeDefaultPrivilegeGrant revokes a default privilege grant for a given object type.
// This is the common pattern used across all default privilege grant resources.
func revokeDefaultPrivilegeGrant(d *schema.ResourceData, meta interface{}, objectType materialize.EntityType) diag.Diagnostics {
	granteeName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, objectType, granteeName, targetName, privilege)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
