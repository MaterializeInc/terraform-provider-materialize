package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var networkPolicySchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the network policy.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"comment": CommentSchema(false),
	"rule": {
		Description: "Rules for the network policy.",
		Type:        schema.TypeSet,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the rule.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"action": {
					Description: "The action to take for this rule. Currently only 'allow' is supported.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)
						if v != "allow" {
							errs = append(errs, fmt.Errorf("%q must be 'allow', got: %s", key, v))
						}
						return
					},
				},
				"direction": {
					Description: "The direction of traffic the rule applies to. Currently only 'ingress' is supported.",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)
						if v != "ingress" {
							errs = append(errs, fmt.Errorf("%q must be 'ingress', got: %s", key, v))
						}
						return
					},
				},
				"address": {
					Description: "The CIDR block the rule will be applied to.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
		MaxItems: 25,
	},
	"region": RegionSchema(),
}

func NetworkPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "A network policy manages access to the system through IP-based rules.",

		CreateContext: networkPolicyCreate,
		ReadContext:   networkPolicyRead,
		UpdateContext: networkPolicyUpdate,
		DeleteContext: networkPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: networkPolicySchema,
	}
}

func networkPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	policy, err := materialize.ScanNetworkPolicy(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", policy.PolicyName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", policy.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	// Convert rules to terraform format
	ruleList := make([]interface{}, len(policy.Rules))
	for i, r := range policy.Rules {
		ruleList[i] = map[string]interface{}{
			"name":      r.Name,
			"action":    r.Action,
			"direction": r.Direction,
			"address":   r.Address,
		}
	}
	if err := d.Set("rule", ruleList); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func networkPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{
		ObjectType: materialize.NetworkPolicy,
		Name:       name,
	}

	b := materialize.NewNetworkPolicyBuilder(metaDb, o)

	// Convert rules from terraform format
	if v, ok := d.GetOk("rule"); ok {
		rules := v.(*schema.Set).List()
		policyRules := make([]materialize.NetworkPolicyRule, len(rules))
		for i, rule := range rules {
			r := rule.(map[string]interface{})
			policyRules[i] = materialize.NetworkPolicyRule{
				Name:      r["name"].(string),
				Action:    r["action"].(string),
				Direction: r["direction"].(string),
				Address:   r["address"].(string),
			}
		}
		b.Rules(policyRules)
	}

	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	i, err := materialize.NetworkPolicyId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return networkPolicyRead(ctx, d, meta)
}

func networkPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{
		ObjectType: materialize.NetworkPolicy,
		Name:       name,
	}

	if d.HasChange("rule") {
		b := materialize.NewNetworkPolicyBuilder(metaDb, o)
		v := d.Get("rule")
		rules := v.(*schema.Set).List()
		policyRules := make([]materialize.NetworkPolicyRule, len(rules))
		for i, rule := range rules {
			r := rule.(map[string]interface{})
			policyRules[i] = materialize.NetworkPolicyRule{
				Name:      r["name"].(string),
				Action:    r["action"].(string),
				Direction: r["direction"].(string),
				Address:   r["address"].(string),
			}
		}
		b.Rules(policyRules)
		if err := b.Alter(); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return networkPolicyRead(ctx, d, meta)
}

func networkPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{Name: name}
	b := materialize.NewNetworkPolicyBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
