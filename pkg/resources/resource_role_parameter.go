package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var roleParameterSchema = map[string]*schema.Schema{
	"role_name": RoleNameSchema(),
	"variable_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the session variable to modify.",
	},
	"variable_value": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The value to assign to the session variable.",
	},
	"region": RegionSchema(),
}

func RoleParameter() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a system parameter in Materialize.",
		CreateContext: roleParameterCreate,
		ReadContext:   roleParameterRead,
		UpdateContext: roleParameterUpdate,
		DeleteContext: roleParameterDelete,
		Schema:        roleParameterSchema,
	}
}

func roleParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	variableName := d.Get("variable_name").(string)
	variableValue := d.Get("variable_value").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewRoleParameterBuilder(metaDb, roleName, variableName, variableValue)
	if err := b.Set(); err != nil {
		return diag.FromErr(err)
	}

	roleParam := roleName + variableName
	d.SetId(utils.TransformIdWithRegion(string(region), roleParam))

	// TODO: Once possible, implement ReadContext
	// return roleParameterRead(ctx, d, meta)
	return nil
}

// TODO: Once possible, implement ReadContext
func roleParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func roleParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChanges("variable_name", "variable_value") {
		return roleParameterCreate(ctx, d, meta)
	}
	return nil
}

func roleParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleName := d.Get("role_name").(string)
	variableName := d.Get("variable_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewRoleParameterBuilder(metaDb, roleName, variableName, "")
	if err := b.Reset(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
