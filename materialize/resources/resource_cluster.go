package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "A logical cluster, which contains dataflow-powered objects.",

		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		DeleteContext: resourceClusterDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A name for the cluster.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

type ClusterBuilder struct {
	clusterName string
}

func newClusterBuilder(clusterName string) *ClusterBuilder {
	return &ClusterBuilder{
		clusterName: clusterName,
	}
}

func (b *ClusterBuilder) Create() string {
	// Only create empty clusters, manage replicas with separate resource
	return fmt.Sprintf(`CREATE CLUSTER %s REPLICAS ();`, b.clusterName)
}

func (b *ClusterBuilder) Read() string {
	return fmt.Sprintf(`SELECT id, name FROM mz_clusters WHERE name = '%s';`, b.clusterName)
}

func (b *ClusterBuilder) Drop() string {
	return fmt.Sprintf(`DROP CLUSTER %s;`, b.clusterName)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	clusterName := d.Get("name").(string)

	builder := newClusterBuilder(clusterName)
	q := builder.Read()

	var id, name string
	conn.QueryRow(q).Scan(&id, &name)

	d.SetId(id)

	return diags
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	clusterName := d.Get("name").(string)

	builder := newClusterBuilder(clusterName)
	q := builder.Create()

	ExecResource(conn, q)
	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	clusterName := d.Get("name").(string)

	builder := newClusterBuilder(clusterName)
	q := builder.Drop()

	ExecResource(conn, q)
	return diags
}
