package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantSchemaSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("SCHEMA"),
	"schema_name": {
		Description: "The schema that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the schema belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantSchema() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(GrantDefinition, "schema"),

		CreateContext: grantSchemaCreate,
		ReadContext:   grantRead,
		DeleteContext: grantSchemaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSchemaSchema,
	}
}

func grantSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createGrant(ctx, d, meta, "SCHEMA", "schema_name")
}

func grantSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeGrant(d, meta, "SCHEMA", "schema_name")
}
