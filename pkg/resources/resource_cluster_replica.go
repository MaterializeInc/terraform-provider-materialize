package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterReplicaSchema = map[string]*schema.Schema{
	"name":         ObjectNameSchema("replica", true, true),
	"cluster_name": ClusterNameSchema(),
	"size":         SizeSchema("replica", true, true),
	"availability_zone": {
		Description: "The specific availability zone of the replica.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
	},
	"introspection_interval":        IntrospectionIntervalSchema(true, []string{}),
	"introspection_debugging":       IntrospectionDebuggingSchema(true, []string{}),
	"idle_arrangement_merge_effort": IdleArrangementMergeEffortSchema(true, []string{}),
}

func ClusterReplica() *schema.Resource {
	return &schema.Resource{
		Description: "Cluster replicas allocate physical compute resources for a cluster.",

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
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
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

	o := materialize.ObjectSchemaStruct{Name: replicaName, ClusterName: clusterName}
	b := materialize.NewClusterReplicaBuilder(meta.(*sqlx.DB), o)

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
	i, err := materialize.ClusterReplicaId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return clusterReplicaRead(ctx, d, meta)
}

func clusterReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	o := materialize.ObjectSchemaStruct{Name: replicaName, ClusterName: clusterName}
	b := materialize.NewClusterReplicaBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
