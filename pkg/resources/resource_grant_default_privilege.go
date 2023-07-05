package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var grantDefaultPrivilegeSchema = map[string]*schema.Schema{
	"object_type": {
		Description:  "The type of object.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"TABLE", "TYPE", "SECRET", "CONNECTION", "DATABASE", "SCHEMA", "CLUSTER"}, true),
	},
	"grantee_name": {
		Description: "The role name that will gain the default privilege. Use the `PUBLIC` pseudo-role to grant privileges to all roles.",
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
	"database_name": {
		Description: "The default privilege will apply only to objects created in this database, if specified.",
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
	"privilege": {
		Description: "The privilege to grant to the object.",
		Type:        schema.TypeString,
		Required:    true,
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

type DefaultPrivilege struct {
	objectType   string
	granteeId    string
	targetRoleId string
	databaseId   string
	schemaId     string
}

func parseDefaultPrivilegeId(id string) (DefaultPrivilege, error) {
	ie := strings.Split(id, "|")

	if len(ie) != 7 {
		return DefaultPrivilege{}, fmt.Errorf("%s cannot be parsed correctly", id)
	}

	return DefaultPrivilege{
		objectType:   ie[1],
		granteeId:    ie[2],
		targetRoleId: ie[3],
		databaseId:   ie[4],
		schemaId:     ie[5],
	}, nil
}

func grantDefaultPrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	dp, err := parseDefaultPrivilegeId(i)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanDefaultPrivilege(meta.(*sqlx.DB), dp.objectType, dp.granteeId, dp.targetRoleId, dp.databaseId, dp.schemaId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePrivileges(s.Privileges.String)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPrivilege(priviledgeMap[dp.granteeId], privilege) {
		return diag.Errorf("%s: default privilege privilege: %s not set", i, privilege)
	}

	return nil
}

func grantDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	objectType := d.Get("object_type").(string)
	granteenName := d.Get("grantee_name").(string)
	targetRoleName := d.Get("target_role_name").(string)
	databaseName := d.Get("database_name").(string)
	schemaName := d.Get("schema_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), objectType, granteenName, privilege)

	if targetRoleName != "" {
		b.TargetRole(targetRoleName)
	}

	if databaseName != "" {
		b.DatabaseName(databaseName)
	}

	if schemaName != "" {
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

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), objectType, granteenName, privilege)

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
