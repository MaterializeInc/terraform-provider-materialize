package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var systemParameterSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The name of the system parameter.",
	},
	"value": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The value to set for the system parameter.",
	},
	"region": RegionSchema(),
}

func SystemParameter() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a system parameter in Materialize.",
		CreateContext: systemParameterCreate,
		ReadContext:   systemParameterRead,
		UpdateContext: systemParameterUpdate,
		DeleteContext: systemParameterDelete,
		Schema:        systemParameterSchema,
	}
}

func systemParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	paramName := d.Get("name").(string)
	paramValue := d.Get("value").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewSystemParameterBuilder(metaDb, paramName, paramValue)
	if err := b.Set(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), paramName))

	return nil
}

func systemParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	paramName := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	paramValue, err := materialize.ShowSystemParameter(metaDb, paramName)
	if err != nil {
		return diag.Errorf("error reading system parameter %s: %s", paramName, err)
	}

	// Update the Terraform state with the retrieved value
	if err := d.Set("name", paramName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("value", paramValue); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), paramName))

	return nil
}
func systemParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Updates are handled by re-setting the parameter value.
	return systemParameterCreate(ctx, d, meta)
}

func systemParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	paramName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	b := materialize.NewSystemParameterBuilder(metaDb, paramName, "")
	if err := b.Reset(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
