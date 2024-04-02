package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var clusterReplicaSchema = map[string]*schema.Schema{
	"name":         ObjectNameSchema("replica", true, true),
	"cluster_name": ClusterNameSchema(),
	"comment":      CommentSchema(false),
	"size":         SizeSchema("replica", true, true),
	"disk":         DiskSchema(true),
	"availability_zone": {
		Description: "The specific availability zone of the replica.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
	},
	"introspection_interval":        IntrospectionIntervalSchema(true, []string{}),
	"introspection_debugging":       IntrospectionDebuggingSchema(true, []string{}),
	"region":                        RegionSchema(),
}

func ClusterReplica() *schema.Resource {
	return &schema.Resource{
		Description: "Cluster replicas allocate physical compute resources for a cluster.",

		CreateContext: clusterReplicaCreate,
		ReadContext:   clusterReplicaRead,
		UpdateContext: clusterReplicaUpdate,
		DeleteContext: clusterReplicaDelete,

		DeprecationMessage: "Cluster replicas are deprecated. We recommend migrating to a managed cluster using the `materialize_cluster` resource and selecting `size`.",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: clusterReplicaSchema,
	}
}

func clusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanClusterReplica(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.ReplicaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("disk", s.Disk.Bool); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("availability_zone", s.AvailabilityZone.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func clusterReplicaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{
		ObjectType:  "CLUSTER REPLICA",
		Name:        replicaName,
		ClusterName: clusterName,
	}
	b := materialize.NewClusterReplicaBuilder(metaDb, o)

	if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
	}

	if v, ok := d.GetOk("disk"); ok {
		b.Disk(v.(bool))
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

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
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
	i, err := materialize.ClusterReplicaId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return clusterReplicaRead(ctx, d, meta)
}

func clusterReplicaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{
		ObjectType:  "CLUSTER REPLICA",
		Name:        replicaName,
		ClusterName: clusterName,
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return clusterReplicaRead(ctx, d, meta)
}

func clusterReplicaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	replicaName := d.Get("name").(string)
	clusterName := d.Get("cluster_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: replicaName, ClusterName: clusterName}
	b := materialize.NewClusterReplicaBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
