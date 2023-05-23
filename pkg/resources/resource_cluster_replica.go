package resources

import (
	"context"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterReplicaSchema = map[string]*schema.Schema{
	"name": NameSchema("replica", true, true),
	"cluster_name": {
		Description: "The cluster whose resources you want to create an additional computation of.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"size": SizeSchema("replica"),
	"availability_zone": {
		Description: "If you want the replica to reside in a specific availability zone.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"introspection_interval": {
		Description: "The interval at which to collect introspection data.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     "1s",
	},
	"introspection_debugging": {
		Description: "Whether to introspect the gathering of the introspection data.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	},
	"idle_arrangement_merge_effort": {
		Description: "The amount of effort the replica should exert on compacting arrangements during idle periods. This is an unstable option! It may be changed or removed at any time.",
		Type:        schema.TypeInt,
		Optional:    true,
		ForceNew:    true,
	},
}

func ClusterReplica() *schema.Resource {
	return &schema.Resource{
		Description: "A cluster replica is the physical resource which maintains dataflow-powered objects.",

		CreateContext: clusterReplicaCreate,
		ReadContext:   clusterReplicaRead,
		DeleteContext: clusterReplicaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: clusterReplicaSchema,
	}
}

func clusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	s, err := materialize.ScanClusterReplica(meta.(*sqlx.DB), i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ReplicaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("availability_zone", s.AvailabilityZone.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	b := materialize.NewClusterReplicaBuilder(meta.(*sqlx.DB), replicaName, clusterName)

	if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		b.AvailabilityZone(v.(string))
	}

	if v, ok := d.GetOk("introspection_interval"); ok {
		b.IntrospectionInterval(v.(string))
	}

	if v, ok := d.GetOk("introspection_debugging"); ok && v.(bool) {
		b.IntrospectionDebugging()
	}

	if v, ok := d.GetOk("idle_arrangement_merge_effort"); ok {
		b.IdleArrangementMergeEffort(v.(int))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.ClusterReplicaId(meta.(*sqlx.DB), replicaName, clusterName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return clusterReplicaRead(ctx, d, meta)
}

func clusterReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	b := materialize.NewClusterReplicaBuilder(meta.(*sqlx.DB), replicaName, clusterName)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
