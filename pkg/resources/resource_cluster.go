package resources

import (
	"context"
	"database/sql"
	"fmt"
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
	"introspection_interval":  IntrospectionIntervalSchema(false, []string{"size"}),
	"introspection_debugging": IntrospectionDebuggingSchema(false, []string{"size"}),
	"scheduling": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		Description:   "Defines the scheduling parameters for the cluster.",
		RequiredWith:  []string{"size"},
		ConflictsWith: []string{"replication_factor"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"on_refresh": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Configuration for refreshing the cluster.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Enable scheduling to refresh the cluster.",
							},
							"hydration_time_estimate": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Estimated time to hydrate the cluster during refresh.",
							},
							"rehydration_time_estimate": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Estimated time to rehydrate the cluster during refresh. This field is deprecated and will be removed in a future release. Use `hydration_time_estimate` instead.",
								Deprecated:  "Use `hydration_time_estimate` instead.",
							},
						},
					},
				},
			},
		},
	},
	"identify_by_name": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Use the cluster name as the resource identifier in your state file, rather than the internal cluster ID.",
	},
	"region": RegionSchema(),
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "Clusters describe logical compute resources that can be used by sources, sinks, indexes, and materialized views. Managed clusters are created by setting the `size` attribute",

		CreateContext: clusterCreate,
		ReadContext:   clusterRead,
		UpdateContext: clusterUpdate,
		DeleteContext: clusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: clusterImport,
		},

		Schema: clusterSchema,
	}
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	fullId := d.Id()
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	idType := utils.ExtractIdType(fullId)
	value := utils.ExtractId(fullId)
	useNameAsId := d.Get("identify_by_name").(bool)

	s, err := materialize.ScanCluster(metaDb, value, idType == "name")
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	if useNameAsId {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", s.ClusterName.String))
	} else {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "id", s.ClusterId.String))
	}

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

		if v, ok := d.GetOk("availability_zones"); ok && len(v.([]interface{})) > 0 {
			f, err := materialize.GetSliceValueString("availability_zones", v.([]interface{}))
			if err != nil {
				return diag.FromErr(err)
			}
			b.AvailabilityZones(f)
		}

		if v, ok := d.GetOk("introspection_interval"); ok {
			b.IntrospectionInterval(v.(string))
		}

		if v, ok := d.GetOk("introspection_debugging"); ok && v.(bool) {
			b.IntrospectionDebugging()
		}

		if v, ok := d.GetOk("scheduling"); ok {
			b.Scheduling(v.([]interface{}))
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
	identifyByName := d.Get("identify_by_name").(bool)
	var idType, value string
	if identifyByName {
		idType = "name"
		value = clusterName
	} else {
		idType = "id"
		clusterId, err := materialize.ClusterId(metaDb, o)
		if err != nil {
			return diag.FromErr(err)
		}
		value = clusterId
	}
	d.SetId(utils.TransformIdWithTypeAndRegion(string(region), idType, value))

	return clusterRead(ctx, d, meta)
}

func clusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterName := d.Get("name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CLUSTER", Name: clusterName}

	if d.HasChange("identify_by_name") {
		_, newIdentifyByName := d.GetChange("identify_by_name")
		identifyByName := newIdentifyByName.(bool)

		// Get the current ID and extract the value
		fullId := d.Id()
		currentValue := utils.ExtractId(fullId)

		var newIdType, newValue string
		if identifyByName {
			newIdType = "name"
			newValue = clusterName
		} else {
			newIdType = "id"
			clusterId, err := materialize.ClusterId(metaDb, o)
			if err != nil {
				return diag.FromErr(err)
			}
			newValue = clusterId
		}

		if currentValue != newValue {
			newId := utils.TransformIdWithTypeAndRegion(string(region), newIdType, newValue)
			d.SetId(newId)
		}
	}

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
			azs, err := materialize.GetSliceValueString("availability_zones", n.([]interface{}))
			if err != nil {
				return diag.FromErr(err)
			}
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

		if d.HasChange("scheduling") {
			o, n := d.GetChange("scheduling")
			if err := b.SetSchedulingConfig(n); err != nil {
				d.Set("scheduling", o)
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

func clusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return nil, err
	}

	fullId := d.Id()
	idType := utils.ExtractIdType(fullId)
	value := utils.ExtractId(fullId)
	identifyByName := idType == "name"

	s, err := materialize.ScanCluster(metaDb, value, identifyByName)
	if err != nil {
		return nil, fmt.Errorf("error importing cluster %s: %s", fullId, err)
	}

	d.Set("identify_by_name", identifyByName)

	if identifyByName {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "name", s.ClusterName.String))
	} else {
		d.SetId(utils.TransformIdWithTypeAndRegion(string(region), "id", s.ClusterId.String))
	}

	d.Set("name", s.ClusterName.String)
	d.Set("ownership_role", s.OwnerName.String)
	d.Set("replication_factor", s.ReplicationFactor.Int64)
	d.Set("size", s.Size.String)
	d.Set("disk", s.Disk.Bool)
	d.Set("availability_zones", s.AvailabilityZones)
	d.Set("comment", s.Comment.String)

	return []*schema.ResourceData{d}, nil
}
