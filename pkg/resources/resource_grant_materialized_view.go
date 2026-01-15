package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantMaterializedViewSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("MATERIALIZED VIEW"),
	"materialized_view_name": {
		Description: "The materialized view that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The schema that the materialized view being to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the materialized view belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantMaterializedView() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(GrantDefinition, "materialized view"),

		CreateContext: grantMaterializedViewCreate,
		ReadContext:   grantRead,
		DeleteContext: grantMaterializedViewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantMaterializedViewSchema,
	}
}

func grantMaterializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createGrant(ctx, d, meta, "MATERIALIZED VIEW", "materialized_view_name")
}

func grantMaterializedViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeGrant(d, meta, "MATERIALIZED VIEW", "materialized_view_name")
}
