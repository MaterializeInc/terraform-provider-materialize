package resources

import (
	"context"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var clusterReplicaSchema = map[string]*schema.Schema{
	"name": {
		Description: "A name for this replica.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"cluster_name": {
		Description: "The cluster whose resources you want to create an additional computation of.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"size": {
		Description:  "The size of the replica.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(append(replicaSizes, localSizes...), true),
	},
	"availability_zone": {
		Description:  "If you want the replica to reside in a specific availability zone.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(regions, true),
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
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadClusterReplicaParams(i)

	var name, cluster_name, size, availability_zone *string
	if err := conn.QueryRowx(q).Scan(&name, &cluster_name, &size, &availability_zone); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", cluster_name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", size); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("availability_zone", availability_zone); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	builder := materialize.NewClusterReplicaBuilder(replicaName, clusterName)

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		builder.AvailabilityZone(v.(string))
	}

	if v, ok := d.GetOk("introspection_interval"); ok {
		builder.IntrospectionInterval(v.(string))
	}

	if v, ok := d.GetOk("introspection_debugging"); ok && v.(bool) {
		builder.IntrospectionDebugging()
	}

	if v, ok := d.GetOk("idle_arrangement_merge_effort"); ok {
		builder.IdleArrangementMergeEffort(v.(int))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "cluster replica"); err != nil {
		return diag.FromErr(err)
	}
	return clusterReplicaRead(ctx, d, meta)
}

func clusterReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	q := materialize.NewClusterReplicaBuilder(replicaName, clusterName).Drop()

	if err := dropResource(conn, d, q, "cluster replica"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
