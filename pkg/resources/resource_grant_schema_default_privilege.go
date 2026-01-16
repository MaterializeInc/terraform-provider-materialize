package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantSchemaDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name":     GranteeNameSchema(),
	"target_role_name": TargetRoleNameSchema(),
	"database_name":    GrantDefaultDatabaseNameSchema(),
	"privilege":        PrivilegeSchema("SCHEMA"),
	"region":           RegionSchema(),
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
	return createDefaultPrivilegeGrant(ctx, d, meta, "SCHEMA")
}

func grantSchemaDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeDefaultPrivilegeGrant(d, meta, "SCHEMA")
}
