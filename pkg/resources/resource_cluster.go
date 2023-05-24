package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterSchema = map[string]*schema.Schema{
	"name": NameSchema("cluster", true, true),
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "A logical cluster, which contains dataflow-powered objects.",

		CreateContext: clusterCreate,
		ReadContext:   clusterRead,
		DeleteContext: clusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: clusterSchema,
	}
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	s, err := materialize.ScanCluster(meta.(*sqlx.DB), i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), clusterName)

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.ClusterId(meta.(*sqlx.DB), clusterName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), clusterName)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
