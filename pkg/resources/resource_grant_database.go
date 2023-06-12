package resources

import (
	"context"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantDatabaseSchema = map[string]*schema.Schema{
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
		ValidateFunc: validPrivileges("DATABASE"),
	},
	"object": {
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The database name.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		Required:    true,
		MinItems:    1,
		MaxItems:    1,
		ForceNew:    true,
		Description: "The object that is being granted on.",
	},
}

func GrantDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the privileges on a Materailize database for roles.",

		CreateContext: grantDatabaseCreate,
		ReadContext:   grantDatabaseRead,
		DeleteContext: grantDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantDatabaseSchema,
	}
}

func grantDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	ie := strings.Split(i, "|")
	objType := ie[1]
	objId := ie[2]
	roleId := ie[3]

	s, err := materialize.ScanPrivileges(meta.(*sqlx.DB), objType, objId)
	if err != nil {
		return diag.FromErr(err)
	}

	priviledgeMap := materialize.ParsePriviledges(s)
	privilege := d.Get("privilege").(string)

	if !materialize.HasPriviledge(priviledgeMap[roleId], privilege) {
		return diag.Errorf("object does not contain privilege: %s", privilege)
	}

	return nil
}

func grantDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	o := d.Get("object").([]interface{})[0].(map[string]interface{})

	obj := materialize.PriviledgeObjectStruct{
		Type: "DATABASE",
		Name: o["name"].(string),
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

	return grantDatabaseRead(ctx, d, meta)
}

func grantDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	o := d.Get("object").([]interface{})[0].(map[string]interface{})

	b := materialize.NewPrivilegeBuilder(
		meta.(*sqlx.DB),
		roleName,
		privilege,
		materialize.PriviledgeObjectStruct{
			Type: "DATABASE",
			Name: o["name"].(string),
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
