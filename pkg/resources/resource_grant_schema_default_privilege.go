package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantSchemaDefaultPrivilegeSchema = map[string]*schema.Schema{
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
		Computed:    true,
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
		Description:  "The privilege to grant to the object.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validPrivileges("SCHEMA"),
	},
}

func GrantSchemaDefaultPrivilege() *schema.Resource {
	return &schema.Resource{
		Description: "Defines default privileges that will be applied to objects created in the future. It does not affect any existing objects.",

		CreateContext: grantSchemaDefaultPrivilegeCreate,
		ReadContext:   grantDefaultPrivilegeRead,
		DeleteContext: grantSchemaDefaultPrivilegeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSchemaDefaultPrivilegeSchema,
	}
}

func grantSchemaDefaultPrivilegeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "SCHEMA", granteenName, privilege)

	var targetRole, database, schema string
	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		targetRole = v.(string)
		b.TargetRole(targetRole)
	}

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

	// set id
	i, err := materialize.DefaultPrivilegeId(meta.(*sqlx.DB), "SCHEMA", granteenName, targetRole, database, schema, privilege)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantDefaultPrivilegeRead(ctx, d, meta)
}

func grantSchemaDefaultPrivilegeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	granteenName := d.Get("grantee_name").(string)
	privilege := d.Get("privilege").(string)

	b := materialize.NewDefaultPrivilegeBuilder(meta.(*sqlx.DB), "SCHEMA", granteenName, privilege)

	if v, ok := d.GetOk("target_role_name"); ok && v.(string) != "" {
		b.TargetRole(v.(string))
	}

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
