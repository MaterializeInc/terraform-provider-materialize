package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantDatabaseSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("DATABASE"),
	"database_name": {
		Description: "The database that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func GrantDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the privileges on a Materailize database for roles.",

		CreateContext: grantDatabaseCreate,
		ReadContext:   grantRead,
		DeleteContext: grantDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantDatabaseSchema,
	}
}

func grantDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	databaseName := d.Get("database_name").(string)

	obj := materialize.ObjectSchemaStruct{
		ObjectType: "DATABASE",
		Name:       databaseName,
	}

	b := materialize.NewPrivilegeBuilder(meta.(*sqlx.DB), roleName, privilege, obj)

	// grant resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// set grant id
	roleId, err := materialize.RoleId(meta.(*sqlx.DB), roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	i, err := materialize.ObjectId(meta.(*sqlx.DB), obj)
	if err != nil {
		return diag.FromErr(err)
	}

	key := b.GrantKey(i, roleId, privilege)
	d.SetId(key)

	return grantRead(ctx, d, meta)
}

func grantDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewPrivilegeBuilder(
		meta.(*sqlx.DB),
		roleName,
		privilege,
		materialize.ObjectSchemaStruct{
			ObjectType: "DATABASE",
			Name:       databaseName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
