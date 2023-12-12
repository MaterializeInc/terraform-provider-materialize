package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantConnectionSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"privilege": PrivilegeSchema("CONNECTION"),
	"connection_name": {
		Description: "The connection that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The schema that the connection being to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the connection belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"region": RegionSchema(),
}

func GrantConnection() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf(GrantDefinition, "connection"),

		CreateContext: grantConnectionCreate,
		ReadContext:   grantRead,
		DeleteContext: grantConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantConnectionSchema,
	}
}

func grantConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	connectionName := d.Get("connection_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	obj := materialize.MaterializeObject{
		ObjectType:   "CONNECTION",
		Name:         connectionName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
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

func grantConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	connectionName := d.Get("connection_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(
		metaDb,
		roleName,
		privilege,
		materialize.MaterializeObject{
			ObjectType:   "CONNECTION",
			Name:         connectionName,
			SchemaName:   schemaName,
			DatabaseName: databaseName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
