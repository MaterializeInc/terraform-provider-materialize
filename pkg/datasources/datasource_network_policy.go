package datasources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NetworkPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: networkPolicyRead,
		Description: "A network policy data source. This can be used to get information about all network policies in Materialize.",
		Schema: map[string]*schema.Schema{
			"network_policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The network policies in the account",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "The ID of the network policy.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the network policy.",
							Computed:    true,
						},
						"comment": {
							Type:        schema.TypeString,
							Description: "The comment of the network policy.",
							Computed:    true,
						},
						"rules": {
							Type:        schema.TypeList,
							Description: "Rules for the network policy.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the rule.",
										Computed:    true,
									},
									"action": {
										Type:        schema.TypeString,
										Description: "The action to take for this rule. Currently only 'allow' is supported.",
										Computed:    true,
									},
									"direction": {
										Type:        schema.TypeString,
										Description: "The direction of traffic the rule applies to. Currently only 'ingress' is supported.",
										Computed:    true,
									},
									"address": {
										Type:        schema.TypeString,
										Description: "The CIDR block the rule will be applied to.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"region": RegionSchema(),
		},
	}
}

func networkPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Add List function to materialize package
	networkPolicy, err := materialize.ListNetworkPolicies(metaDb)
	if err != nil {
		return diag.FromErr(err)
	}

	networkPolicyFormatted := []map[string]interface{}{}
	for _, p := range networkPolicy {
		networkPolicyMap := map[string]interface{}{}

		networkPolicyMap["id"] = p.PolicyId.String
		networkPolicyMap["name"] = p.PolicyName.String
		networkPolicyMap["comment"] = p.Comment.String

		// Format rules
		rules := []map[string]interface{}{}
		for _, r := range p.Rules {
			ruleMap := map[string]interface{}{
				"name":      r.Name,
				"action":    r.Action,
				"direction": r.Direction,
				"address":   r.Address,
			}
			rules = append(rules, ruleMap)
		}
		networkPolicyMap["rules"] = rules

		networkPolicyFormatted = append(networkPolicyFormatted, networkPolicyMap)
	}

	if err := d.Set("network_policies", networkPolicyFormatted); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), "network_policies"))
	return diags
}
