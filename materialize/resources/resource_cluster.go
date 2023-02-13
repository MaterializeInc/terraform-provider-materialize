package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var clusterSchema = map[string]*schema.Schema{
	"name": {
		Description: "A name for the cluster.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func Cluster() *schema.Resource {
	return &schema.Resource{
		Description: "A logical cluster, which contains dataflow-powered objects.",

		CreateContext: clusterCreate,
		ReadContext:   clusterRead,
		DeleteContext: clusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: clusterSchema,
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

func (b *ClusterBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_clusters WHERE name = '%s';`, b.clusterName)
}

func (b *ClusterBuilder) Drop() string {
	return fmt.Sprintf(`DROP CLUSTER %s;`, b.clusterName)
}

func readClusterParams(id string) string {
	return fmt.Sprintf("SELECT name FROM mz_clusters WHERE id = '%s';", id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _cluster struct {
	name sql.NullString `db:"name"`
}

func clusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readClusterParams(i)

	readResource(conn, d, i, q, _cluster{}, "cluster")
	return nil
}

func clusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	clusterName := d.Get("name").(string)

	builder := newClusterBuilder(clusterName)
	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "cluster")
	return clusterRead(ctx, d, meta)
}

func clusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	clusterName := d.Get("name").(string)

	builder := newClusterBuilder(clusterName)
	q := builder.Drop()

	dropResource(conn, d, q, "cluster")
	return nil
}
