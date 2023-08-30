package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantSchemaDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name":     GranteeNameSchema(),
	"target_role_name": TargetRoleNameSchema(),
	"database_name":    GrantDefaultDatabaseNameSchema(),
	"privilege":        PrivilegeSchema("SCHEMA"),
}

func GrantSchemaDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: DefaultPrivilegeDefinition,

		CreateContext: grantSchemaDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantSchemaDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSchemaDefaultPrivilegeSchema,
	}
}

func grantSchemaDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteeName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "SCHEMA", granteeName, targetName, privilege)

	var database string
	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		database = v.(string)
		b.DatabaseName(database)
	}

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// Query ids
	gId, err := materialize.RoleId(meta.(*sqlx.DB), granteeName)
	if err != nil {
		return diag.FromErr(err)
	}

	tId, err := materialize.RoleId(meta.(*sqlx.DB), targetName)
	if err != nil {
		return diag.FromErr(err)
	}

	var dId string
	if database != "" {
		dId, err = materialize.DatabaseId(meta.(*sqlx.DB), materialize.ObjectSchemaStruct{Name: database})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	key := b.GrantKey("SCHEMA", gId, tId, dId, "", privilege)
	d.SetId(key)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantSchemaDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "SCHEMA", granteenName, targetName, privilege)

	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		b.DatabaseName(v.(string))
	}

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
