package resources

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	"availability_zones": {
		Description:  "The specific availability zones of the cluster.",
		Type:         schema.TypeList,
		Elem:         &schema.Schema{Type: schema.TypeString},
		Computed:     true,
		Optional:     true,
		RequiredWith: []string{"size"},
	},
	"introspection_interval":        IntrospectionIntervalSchema(false, []string{"size"}),
	"introspection_debugging":       IntrospectionDebuggingSchema(false, []string{"size"}),
	"region":                        RegionSchema(),
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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanCluster(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

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

	if err := d.Set("availability_zones", s.AvailabilityZones); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}
	b := materialize.NewClusterBuilder(metaDb, o)

	// managed cluster options
	if size, ok := d.GetOk("size"); ok {
		b.Size(size.(string))

		if v, ok := d.GetOkExists("replication_factor"); ok {
			r := v.(int)
			b.ReplicationFactor(&r)
		}

		// DISK option not supported for cluster sizes ending in cc or C because disk is always enabled
		if strings.HasSuffix(size.(string), "cc") || strings.HasSuffix(size.(string), "C") {
			log.Printf("[WARN] disk option not supported for cluster size %s, disk is always enabled", size)
			d.Set("disk", true)
		} else {
			if v, ok := d.GetOk("disk"); ok {
				b.Disk(v.(bool))
			}
		}

		if v, ok := d.GetOk("availability_zones"); ok {
			f := materialize.GetSliceValueString(v.([]interface{}))
			b.AvailabilityZones(f)
		}

		if v, ok := d.GetOk("introspection_interval"); ok {
			b.IntrospectionInterval(v.(string))
		}

		if v, ok := d.GetOk("introspection_debugging"); ok && v.(bool) {
			b.IntrospectionDebugging()
		}
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.ClusterId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return clusterRead(ctx, d, meta)
}

func clusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)
		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	b := materialize.NewClusterBuilder(metaDb, o)
	if _, ok := d.GetOk("size"); ok {
		if d.HasChange("size") {
			_, newSize := d.GetChange("size")
			if err := b.Resize(newSize.(string)); err != nil {
				return diag.FromErr(err)
			}

		}

		if d.HasChange("disk") {
			// DISK option not supported for cluster sizes ending in cc or C because disk is always enabled
			size := d.Get("size").(string)
			if strings.HasSuffix(size, "cc") || strings.HasSuffix(size, "C") {
				log.Printf("[WARN] disk option not supported for cluster size %s, disk is always enabled", size)
				d.Set("disk", true)
			} else {
				_, newDisk := d.GetChange("disk")
				if err := b.SetDisk(newDisk.(bool)); err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if d.HasChange("replication_factor") {
			_, n := d.GetChange("replication_factor")
			if err := b.SetReplicationFactor(n.(int)); err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("availability_zones") {
			_, n := d.GetChange("availability_zones")
			azs := materialize.GetSliceValueString(n.([]interface{}))
			b := materialize.NewClusterBuilder(metaDb, o)
			if err := b.SetAvailabilityZones(azs); err != nil {
				return diag.FromErr(err)
			}
		}

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
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: clusterName}
	b := materialize.NewClusterBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
