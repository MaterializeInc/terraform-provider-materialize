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
	"name":           ObjectNameSchema("cluster", true, true),
	"comment":        CommentSchema(false),
	"ownership_role": OwnershipRoleSchema(),
	"size":           SizeSchema("managed cluster", false, false),
	"replication_factor": {
		Description:  "The number of replicas of each dataflow-powered object to maintain.",
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		RequiredWith: []string{"size"},
	},
	"disk": DiskSchema(false),
	// "availability_zones": {
	// 	Description: "The specific availability zones of the cluster.",
	// 	Type:        schema.TypeList,
	// 	Elem:        &schema.Schema{Type: schema.TypeString},
	// 	Computed:    true,
	// 	RequiredWith: []string{"size"},
	// },
	"introspection_interval":        IntrospectionIntervalSchema(false, []string{"size"}),
	"introspection_debugging":       IntrospectionDebuggingSchema(false, []string{"size"}),
	"idle_arrangement_merge_effort": IdleArrangementMergeEffortSchema(false, []string{"size"}),
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "Clusters describe logical compute resources that can be used by sources, sinks, indexes, and materialized views. Managed clusters are created by setting the `size` attribute",

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

	if err := d.Set("disk", s.Disk.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}
	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)

	// managed cluster options
	if size, ok := d.GetOk("size"); ok {
		b.Size(size.(string))

		if v, ok := d.GetOkExists("replication_factor"); ok {
			r := v.(int)
			b.ReplicationFactor(&r)
		}

		if v, ok := d.GetOk("disk"); ok {
			b.Disk(v.(bool))
		}

		// TODO: Disable until supported on create
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

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
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

	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), o)
		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)
	if _, ok := d.GetOk("size"); ok {
		if d.HasChange("size") {
			_, newSize := d.GetChange("size")
			if err := b.Resize(newSize.(string)); err != nil {
				return diag.FromErr(err)
			}

		}

		if d.HasChange("disk") {
			_, newDisk := d.GetChange("disk")
			if err := b.SetDisk(newDisk.(bool)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("replication_factor") {
			_, n := d.GetChange("replication_factor")
			if err := b.SetReplicationFactor(n.(int)); err != nil {
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

		if d.HasChange("introspection_interval") {
			_, n := d.GetChange("introspection_interval")
			if err := b.SetIntrospectionInterval(n.(string)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("introspection_debugging") {
			_, n := d.GetChange("introspection_debugging")
			if err := b.SetIntrospectionDebugging(n.(bool)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("idle_arrangement_merge_effort") {
			_, n := d.GetChange("idle_arrangement_merge_effort")
			if err := b.SetIdleArrangementMergeEffort(n.(int)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(meta.(*sqlx.DB), o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	o := materialize.MaterializeObject{Name: clusterName}
	b := materialize.NewClusterBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
