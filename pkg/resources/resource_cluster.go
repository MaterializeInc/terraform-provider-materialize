package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterSchema = map[string]*schema.Schema{
	"name":           NameSchema("cluster", true, true),
	"ownership_role": OwnershipRole(),
	"replication_factor": {
		Description: "The number of replicas of each dataflow-powered object to maintain.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
	"size": SizeSchema("cluster"),
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "A logical cluster, which contains dataflow-powered objects.",

		CreateContext: clusterCreate,
		ReadContext:   clusterRead,
		UpdateContext: clusterUpdate,
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
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("replication_factor", s.ReplicationFactor.Int64); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	o := materialize.ObjectSchemaStruct{Name: clusterName}
	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)

	// replication factor
	if v, ok := d.GetOk("replication_factor"); ok {
		b.ReplicationFactor(v.(int))

		// size
		if v, ok := d.GetOk("size"); ok {
			b.Size(v.(string))
		}
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "CLUSTER", o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.ClusterId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return clusterRead(ctx, d, meta)
}

func clusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")

		o := materialize.ObjectSchemaStruct{Name: clusterName}
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "CLUSTER", o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	o := materialize.ObjectSchemaStruct{Name: clusterName}
	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
