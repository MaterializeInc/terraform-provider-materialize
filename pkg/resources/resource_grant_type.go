package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantTypeSchema = map[string]*schema.Schema{
	"role_name": {
		Description: "The name of the role to grant privilege to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"privilege": {
		Description:  "The privilege to grant to the object.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validPrivileges("TYPE"),
	},
	"type_name": {
		Description: "The type that is being granted on.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"schema_name": {
		Description: "The schema that the type being to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database that the type belongs to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func GrantType() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the privileges on a Materailize type for roles.",

		CreateContext: grantTypeCreate,
		ReadContext:   grantRead,
		DeleteContext: grantTypeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantTypeSchema,
	}
}

func grantTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	typeName := d.Get("type_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	obj := materialize.PriviledgeObjectStruct{
		Type:         "TYPE",
		Name:         typeName,
		SchemaName:   schemaName,
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

	i, err := materialize.PrivilegeId(meta.(*sqlx.DB), obj, roleId, privilege)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return grantRead(ctx, d, meta)
}

func grantTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)
	typeName := d.Get("type_name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewPrivilegeBuilder(
		meta.(*sqlx.DB),
		roleName,
		privilege,
		materialize.PriviledgeObjectStruct{
			Type:         "TYPE",
			Name:         typeName,
			SchemaName:   schemaName,
			DatabaseName: databaseName,
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
