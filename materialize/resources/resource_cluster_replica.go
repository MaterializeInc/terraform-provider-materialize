package resources

import (
	"context"
	"fmt"
	"strings"

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

type ClusterReplicaBuilder struct {
	replicaName                string
	clusterName                string
	size                       string
	availabilityZone           string
	introspectionInterval      string
	introspectionDebugging     bool
	idleArrangementMergeEffort int
}

func newClusterReplicaBuilder(replicaName, clusterName string) *ClusterReplicaBuilder {
	return &ClusterReplicaBuilder{
		replicaName: replicaName,
		clusterName: clusterName,
	}
}

func (b *ClusterReplicaBuilder) Size(s string) *ClusterReplicaBuilder {
	b.size = s
	return b
}

func (b *ClusterReplicaBuilder) AvailabilityZone(z string) *ClusterReplicaBuilder {
	b.availabilityZone = z
	return b
}

func (b *ClusterReplicaBuilder) IntrospectionInterval(i string) *ClusterReplicaBuilder {
	b.introspectionInterval = i
	return b
}

func (b *ClusterReplicaBuilder) IntrospectionDebugging() *ClusterReplicaBuilder {
	b.introspectionDebugging = true
	return b
}

func (b *ClusterReplicaBuilder) IdleArrangementMergeEffort(e int) *ClusterReplicaBuilder {
	b.idleArrangementMergeEffort = e
	return b
}

func (b *ClusterReplicaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CLUSTER REPLICA %s.%s`, b.clusterName, b.replicaName))

	var p []string
	if b.size != "" {
		s := fmt.Sprintf(` SIZE = '%s'`, b.size)
		p = append(p, s)
	}

	if b.availabilityZone != "" {
		a := fmt.Sprintf(` AVAILABILITY ZONE = '%s'`, b.availabilityZone)
		p = append(p, a)
	}

	if b.introspectionInterval != "" {
		i := fmt.Sprintf(` INTROSPECTION INTERVAL = '%s'`, b.introspectionInterval)
		p = append(p, i)
	}

	if b.introspectionDebugging {
		p = append(p, ` INTROSPECTION DEBUGGING = TRUE`)
	}

	if b.idleArrangementMergeEffort != 0 {
		m := fmt.Sprintf(` IDLE ARRANGEMENT MERGE EFFORT = %d`, b.idleArrangementMergeEffort)
		p = append(p, m)
	}

	if len(p) > 0 {
		p := strings.Join(p[:], ",")
		q.WriteString(p)
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *ClusterReplicaBuilder) Drop() string {
	return fmt.Sprintf(`DROP CLUSTER REPLICA %s.%s;`, b.clusterName, b.replicaName)
}

func (b *ClusterReplicaBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_cluster_replicas.id
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.name = '%s'
		AND mz_clusters.name = '%s';`, b.replicaName, b.clusterName)
}

func readClusterReplicaParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_cluster_replicas.name,
			mz_clusters.name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.id = '%s';`, id)
}

func clusterReplicaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readClusterReplicaParams(i)

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

	builder := newClusterReplicaBuilder(replicaName, clusterName)

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

	q := newClusterReplicaBuilder(replicaName, clusterName).Drop()

	if err := dropResource(conn, d, q, "cluster replica"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
