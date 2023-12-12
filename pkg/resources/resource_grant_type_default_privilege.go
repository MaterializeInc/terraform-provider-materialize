package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var grantTypeDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name":     GranteeNameSchema(),
	"target_role_name": TargetRoleNameSchema(),
	"database_name":    GrantDefaultDatabaseNameSchema(),
	"schema_name":      GrantDefaultSchemaNameSchema(),
	"privilege":        PrivilegeSchema("TYPE"),
	"region":           RegionSchema(),
}

func GrantTypeDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: DefaultPrivilegeDefinition,

		CreateContext: grantTypeDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantTypeDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantTypeDefaultPrivilegeSchema,
	}
}

func grantTypeDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteeName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, "TYPE", granteeName, targetName, privilege)

	var database, schema string
	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		database = v.(string)
		b.DatabaseName(database)
	}

	if v, ok := d.GetOk("schema_name"); ok && v.(string) != "" {
		schema = v.(string)
		b.SchemaName(schema)
	}

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// Query ids
	gId, err := materialize.RoleId(metaDb, granteeName)
	if err != nil {
		return diag.FromErr(err)
	}

	tId, err := materialize.RoleId(metaDb, targetName)
	if err != nil {
		return diag.FromErr(err)
	}

	var dId, sId string
	if database != "" {
		dId, err = materialize.DatabaseId(metaDb, materialize.MaterializeObject{Name: database})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if schema != "" {
		sId, err = materialize.SchemaId(metaDb, materialize.MaterializeObject{Name: schema, DatabaseName: database})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	key := b.GrantKey(string(region), "TYPE", gId, tId, dId, sId, privilege)
	d.SetId(key)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantTypeDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	targetName := d.Get("target_role_name").(string)
	privilege := d.Get("privilege").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewDefaultPrivilegeBuilder(metaDb, "TYPE", granteenName, targetName, privilege)

	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		b.DatabaseName(v.(string))
	}

	if v, ok := d.GetOk("schema_name"); ok && v.(string) != "" {
		b.SchemaName(v.(string))
	}

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
