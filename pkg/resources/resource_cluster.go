package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Description: "Use the cluster name as the resource identifier in your state file, rather than the internal cluster ID. This is particularly useful in scenarios like dbt-materialize blue/green deployments, where clusters are swapped but the ID changes. By identifying by name, the resource can be managed consistently even when the underlying cluster ID is updated.",
	},
	"region": RegionSchema(),
	"wait_until_ready": {
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Defines the parameters for the WAIT UNTIL READY options",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Enable wait_until_ready.",
				},
				"timeout": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "0s",
					Description:  "Max duration to wait for the new replicas to be ready.",
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^\\d+[smh]{1}$"), "Must be a valid duration in the form of <int><unit> ex: 1s, 10m"),
				},
				"on_timeout": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Action to take on timeout: COMMIT|ROLLBACK",
					Default:      "COMMIT",
					ValidateFunc: validation.StringInSlice([]string{"COMMIT", "ROLLBACK"}, true),
				},
			},
		},
	},
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

	// The disk attr is deprecated and is not configurable
	if err := d.Set("disk", true); err != nil {
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
	o := materialize.MaterializeObject{ObjectType: materialize.Cluster, Name: clusterName}
	b := materialize.NewClusterBuilder(metaDb, o)

	// managed cluster options
	if size, ok := d.GetOk("size"); ok {
		b.Size(size.(string))

		if v, ok := d.GetOkExists("replication_factor"); ok {
			r := v.(int)
			b.ReplicationFactor(&r)
		}

		// TODO: remove this once the disk attr is removed
		// The disk attr is deprecated and is not configurable
		log.Printf("[DEBUG] disk option is deprecated.")

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
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
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
	o := materialize.MaterializeObject{ObjectType: materialize.Cluster, Name: clusterName}

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
	changed := false
	if d.HasChange("size") {
		_, newSize := d.GetChange("size")
		b.SetSize(newSize.(string))
		changed = true
	}

	if d.HasChange("disk") {
		// TODO: remove this once the disk attr is removed
		// The disk attr is deprecated and is not configurable
		log.Printf("[DEBUG] disk option is deprecated and always enabled, ignoring disk change")
	}

	if d.HasChange("replication_factor") {
		_, n := d.GetChange("replication_factor")
		b.SetReplicationFactor(n.(int))
		changed = true
	}

	if d.HasChange("availability_zones") {
		_, n := d.GetChange("availability_zones")
		azs, err := materialize.GetSliceValueString("availability_zones", n.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.SetAvailabilityZones(azs)
		changed = true
	}

	if d.HasChange("introspection_interval") {
		_, n := d.GetChange("introspection_interval")
		b.SetIntrospectionInterval(n.(string))
		changed = true
	}

	if d.HasChange("introspection_debugging") {
		_, n := d.GetChange("introspection_debugging")
		b.SetIntrospectionDebugging(n.(bool))
		changed = true
	}

	if d.HasChange("scheduling") {
		_, n := d.GetChange("scheduling")
		// If the scheduling has changed we need to set that now.
		// There are some conflicting options that require scheduling
		// options to always be adjusted first, enabling a schedule
		// on a cluster with a replication factor will remove it, and
		// you must first set the cluster schedule to manual in order to
		// add a replication factor.
		// Note we don't set this in the `b` ClusterBuilder
		// and we do not set changed here.
		config := materialize.GetSchedulingConfig(n)
		if err := b.AlterClusterScheduling(config); err != nil {
			return diag.FromErr(err)
		}
	}

	if changed {
		_, reconfigOptsRaw := d.GetChange("wait_until_ready")
		reconfigOpts := b.GetReconfigOpts(reconfigOptsRaw)
		if err := b.AlterCluster(reconfigOpts); err != nil {
			return diag.FromErr(err)
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
	// Disk is always enabled for all clusters now (deprecated feature)
	d.Set("disk", true)
	d.Set("availability_zones", s.AvailabilityZones)
	d.Set("comment", s.Comment.String)

	return []*schema.ResourceData{d}, nil
}
