package resources

import (
	"context"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var grantSchema = map[string]*schema.Schema{
	"role_name": {
		Description: "The name of the role to grant privilege to.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"privilege": {
		Description: "The privilege to grant to the object.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"object": {
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Description: "The type of object that is being granted on.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"name": {
					Description: "The object name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"schema_name": {
					Description: "The object schema name if applicable.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"database_name": {
					Description: "The object database name if applicable.",
					Type:        schema.TypeString,
					Optional:    true,
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

func Grant() *schema.Resource {
	return &schema.Resource{
		Description: "Manages the privileges on Materailize objects for roles.",

		CreateContext: grantCreate,
		ReadContext:   grantRead,
		DeleteContext: grantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: grantSchema,
	}
}

func grantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("Object does not contain privilege: %s", privilege)
	}

	return nil
}

func grantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	o := d.Get("object").([]interface{})[0].(map[string]interface{})

	obj := materialize.PriviledgeObjectStruct{
		Type:         o["type"].(string),
		Name:         o["name"].(string),
		SchemaName:   o["schema_name"].(string),
		DatabaseName: o["database_name"].(string),
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

func grantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	privilege := d.Get("privilege").(string)

	o := d.Get("object").([]interface{})[0].(map[string]interface{})

	b := materialize.NewPrivilegeBuilder(
		meta.(*sqlx.DB),
		roleName,
		privilege,
		materialize.PriviledgeObjectStruct{
			Type:         o["type"].(string),
			Name:         o["name"].(string),
			SchemaName:   o["schema_name"].(string),
			DatabaseName: o["database_name"].(string),
		},
	)

	if err := b.Revoke(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
