package resources

import (
	"context"
	"fmt"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantSecretSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("SECRET"),
	"secret_name": {
		Description: "The secret that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The schema that the secret being to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the secret belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantSecret() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(GrantDefinition, "secret"),

		CreateContext: grantSecretCreate,
		ReadContext:   grantRead,
		DeleteContext: grantSecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSecretSchema,
	}
}

func grantSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createGrant(ctx, d, meta, materialize.Secret, "secret_name")
}

func grantSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return revokeGrant(d, meta, materialize.Secret, "secret_name")
}
