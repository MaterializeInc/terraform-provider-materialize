package resources

import (
	"context"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var grantDefaultPrivilegeSchema = map[string]*schema.Schema{
	"grantee_name": {
		Description: "The role name that will gain the default privilege. Use the `PUBLIC` pseudo-role to grant privileges to all roles.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"object_type": {
		Description:  "The type of object.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"TABLE", "TYPE", "SECRET", "CONNECTION", "DATABASE", "SCHEMA", "CLUSTER"}, true),
	},
	"privilege": {
		Description: "The privilege to grant to the object.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"target_role_name": {
		Description: "The default privilege will apply to objects created by this role. If this is left blank, then the current role is assumed. Use the `PUBLIC` pseudo-role to target objects created by all roles. If using `ALL` will apply to objects created by all roles",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The default privilege will apply only to objects created in this schema, if specified.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The default privilege will apply only to objects created in this database, if specified.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
}

func GrantDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: "Defines default privileges that will be applied to objects created in the future. It does not affect any existing objects.",

		CreateContext: grantDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantDefaultPrivilegeSchema,
	}
}

func grantDefaultPrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	ie := strings.Split(i, "|")
	objType := ie[1]
	granteeId := ie[2]
	targetRoleId := ie[3]
	databaseId := ie[4]
	schemaId := ie[5]

	s, err := materialize.ScanDefaultPrivilege(meta.(*sqlx.DB), objType, granteeId, targetRoleId, databaseId, schemaId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePrivileges(s.Privileges.String)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(priviledgeMap[granteeId], privilege) {
		return diag.Errorf("%s: default privilege privilege: %s not set", i, privilege)
	}

	return nil
}

func grantDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	objectType := d.Get("object_type").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), objectType, privilege, granteenName)

	var targetRoleName string
	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		targetRoleName = v.(string)
		b.TargetRole(targetRoleName)
	}

	var databaseName string
	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		databaseName = v.(string)
		b.DatabaseName(databaseName)
	}

	var schemaName string
	if v, ok := d.GetOk("schema_name"); ok && v.(string) != "" {
		schemaName = v.(string)
		b.SchemaName(schemaName)
	}

	// create resource
	if err := b.Grant(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.DefaultPrivilegeId(meta.(*sqlx.DB), objectType, granteenName, targetRoleName, databaseName, schemaName, privilege)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	objectType := d.Get("object_type").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), objectType, privilege, granteenName)

	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		b.TargetRole(v.(string))
	}

	if v, ok := d.GetOk("schema_name"); ok && v.(string) != "" {
		b.SchemaName(v.(string))
	}

	if v, ok := d.GetOk("database_name"); ok && v.(string) != "" {
		b.DatabaseName(v.(string))
	}

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
