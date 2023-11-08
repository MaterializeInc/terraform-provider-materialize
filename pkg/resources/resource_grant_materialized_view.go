package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
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
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	mviewName := d.Get("materialized_view_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	obj := materialize.MaterializeObject{
		ObjectType:   "MATERIALIZED VIEW",
		Name:         mviewName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
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

	key := b.GrantKey(utils.Region, i, roleId, privilege)
	d.SetId(key)

	return grantRead(ctx, d, meta)
}

func grantMaterializedViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	mviewName := d.Get("materialized_view_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewPrivilegeBuilder(
		metaDb,
		roleName,
		privilege,
		materialize.MaterializeObject{
			ObjectType:   "MATERIALIZED VIEW",
			Name:         mviewName,
			SchemaName:   schemaName,
			DatabaseName: databaseName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
