package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
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
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	obj := materialize.MaterializeObject{
		ObjectType:   "SCHEMA",
		Name:         schemaName,
		DatabaseName: databaseName,
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

	key := b.GrantKey(utils.Region, i, roleId, privilege)
	d.SetId(key)

	return grantRead(ctx, d, meta)
}

func grantSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewPrivilegeBuilder(
		meta.(*sqlx.DB),
		roleName,
		privilege,
		materialize.MaterializeObject{
			ObjectType:   "SCHEMA",
			Name:         schemaName,
			DatabaseName: databaseName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
