package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var clusterSchema = map[string]*schema.Schema{
	"name":           NameSchema("cluster", true, true),
	"ownership_role": OwnershipRole(),
	"replication_factor": {
		Description:  "The number of replicas of each dataflow-powered object to maintain.",
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"size"},
	},
	"size": {
		Description:  "The size of the managed cluster.",
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(append(replicaSizes, localSizes...), true),
	},
	// TODO: Enable availability_zones when in system catalog
	// "availability_zones": {
	// 	Description: "If you want the cluster to reside in specific availability zones.",
	// 	Type:        schema.TypeList,
	// 	Elem:        &schema.Schema{Type: schema.TypeString},
	// 	Optional:    true,
	// },
	"introspection_interval": {
		Description: "The interval at which to collect introspection data.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "1s",
	},
	"introspection_debugging": {
		Description:   "Whether to introspect the gathering of the introspection data.",
		Type:          schema.TypeBool,
		Optional:      true,
		Default:       false,
		ConflictsWith: []string{"introspection_interval"},
	},
	"idle_arrangement_merge_effort": {
		Description: "The amount of effort the cluster should exert on compacting arrangements during idle periods. This is an unstable option! It may be changed or removed at any time.",
		Type:        schema.TypeInt,
		Optional:    true,
	},
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

	// size and replication_factor for managed clusters
	if replicationFactor, replicationFactorOk := d.GetOk("replication_factor"); replicationFactorOk {
		if size, sizeOk := d.GetOk("size"); sizeOk {
			b.ReplicationFactor(replicationFactor.(int))
			b.Size(size.(string))

			// if v, ok := d.GetOk("availability_zones"); ok {
			// 	azs := materialize.GetSliceValueString(v.([]interface{}))
			// 	b.AvailabilityZones(azs)
			// }

			if v, ok := d.GetOk("introspection_interval"); ok {
				b.IntrospectionInterval(v.(string))
			}

			if v, ok := d.GetOk("introspection_debugging"); ok && v.(bool) {
				b.IntrospectionDebugging()
			}

			if v, ok := d.GetOk("idle_arrangement_merge_effort"); ok {
				b.IdleArrangementMergeEffort(v.(int))
			}
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

	// Managed Cluster Options
	if _, ok := d.GetOk("size"); ok {
		o := materialize.ObjectSchemaStruct{Name: clusterName}

		if d.HasChange("size") {
			_, newSize := d.GetChange("size")
			b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
			if err := b.Resize(newSize.(string)); err != nil {
				return diag.FromErr(err)
			}

		}

		// if d.HasChange("availability_zones") {
		// 	_, n := d.GetChange("availability_zones")
		// 	azs := materialize.GetSliceValueString(n.([]interface{}))
		// 	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
		// 	if err := b.SetAvailabilityZones(azs); err != nil {
		// 		return diag.FromErr(err)
		// 	}
		// }

		if d.HasChange("replication_factor") {
			_, n := d.GetChange("replication_factor")
			b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
			if err := b.SetReplicationFactor(n.(int)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("introspection_interval") {
			_, n := d.GetChange("introspection_interval")
			b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
			if err := b.SetIntrospectionInterval(n.(string)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("introspection_debugging") {
			_, n := d.GetChange("introspection_debugging")
			b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
			if err := b.SetIntrospectionDebugging(n.(bool)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("idle_arrangement_merge_effort") {
			_, n := d.GetChange("idle_arrangement_merge_effort")
			b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
			if err := b.SetIdleArrangementMergeEffort(n.(int)); err != nil {
				return diag.FromErr(err)
			}
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
