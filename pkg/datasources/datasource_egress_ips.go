package datasources

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func EgressIps() *schema.Resource {
	return &schema.Resource{
		ReadContext: EgressIpsRead,
		Schema: map[string]*schema.Schema{
			"egress_ips": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The egress IPs in the account",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func EgressIpsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metaDb, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	conn := metaDb

	q := materialize.ReadEgressIpsDatasource()

	rows, err := conn.Query(q)

	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("[DEBUG] no egress IPs found in account")
		return diag.FromErr(err)
	} else if err != nil {
		log.Println("[DEBUG] failed to list egress IPs")
		return diag.FromErr(err)
	}

	egressIps := []string{}
	for rows.Next() {
		var egressIp string
		err := rows.Scan(&egressIp)
		if err != nil {
			log.Println("[DEBUG] failed to scan egress IP")
			return diag.FromErr(err)
		}
		egressIps = append(egressIps, egressIp)
	}

	if err := rows.Err(); err != nil {
		log.Println("[DEBUG] failed to list egress IPs")
		return diag.FromErr(err)
	}

	if err := d.Set("egress_ips", egressIps); err != nil {
		log.Println("[DEBUG] failed to set egress_ips")
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion("egress_ips"))

	return diags
}
