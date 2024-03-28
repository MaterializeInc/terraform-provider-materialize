package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/frontegg"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var Scim2GroupSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the SCIM group.",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A description of the SCIM group.",
	},
}

func SCIM2Group() *schema.Resource {
	return &schema.Resource{
		CreateContext: scim2GroupCreate,
		ReadContext:   scim2GroupRead,
		UpdateContext: scim2GroupUpdate,
		DeleteContext: scim2GroupDelete,
		Schema:        Scim2GroupSchema,
		Description:   "The SCIM group resource allows you to manage user groups in Frontegg.",
	}
}

func scim2GroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	params := frontegg.GroupCreateParams{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	group, err := frontegg.CreateSCIMGroup(ctx, client, params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating SCIM group: %s", err))
	}

	d.SetId(group.ID)
	return scim2GroupRead(ctx, d, meta)
}

func scim2GroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	group, err := frontegg.GetSCIMGroupByID(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching SCIM group: %s", err))
	}

	d.Set("name", group.Name)
	d.Set("description", group.Description)

	return nil
}

func scim2GroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	groupID := d.Id()

	// Update group attributes if they have changed
	if d.HasChanges("name", "description") {
		params := frontegg.GroupUpdateParams{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}

		err := frontegg.UpdateSCIMGroup(ctx, client, groupID, params)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating SCIM group: %s", err))
		}
	}

	return scim2GroupRead(ctx, d, meta)
}

func scim2GroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerMeta, err := utils.GetProviderMeta(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	client := providerMeta.Frontegg

	err = frontegg.DeleteSCIMGroup(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCIM group: %s", err))
	}

	d.SetId("")
	return nil
}
