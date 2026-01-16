package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantViewSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("VIEW"),
	"view_name": {
		Description: "The view that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The schema that the view being to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the view belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantView() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the privileges on a Materailize view for roles.",

		CreateContext: grantViewCreate,
		ReadContext:   grantRead,
		DeleteContext: grantViewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantViewSchema,
	}
}

func grantViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createGrant(ctx, d, meta, "VIEW", "view_name")
}

func grantViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeGrant(d, meta, "VIEW", "view_name")
}
