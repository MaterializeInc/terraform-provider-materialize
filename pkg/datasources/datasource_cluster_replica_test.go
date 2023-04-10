package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestClusterReplicaDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, ClusterReplica().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"id", "name", "cluster", "size", "availability_zone"}).
			AddRow("u1", "replica", "cluster", "small", "us-east-1")
		mock.ExpectQuery(`
			SELECT
				mz_cluster_replicas.id,
				mz_cluster_replicas.name,
				mz_clusters.name,
				mz_cluster_replicas.size,
				mz_cluster_replicas.availability_zone
			FROM mz_cluster_replicas
			JOIN mz_clusters
				ON mz_cluster_replicas.cluster_id = mz_clusters.id;`).WillReturnRows(ir)

		if err := clusterReplicaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
