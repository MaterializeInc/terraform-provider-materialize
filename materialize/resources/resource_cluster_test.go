package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{"name": "cluster"}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`CREATE CLUSTER cluster REPLICAS \(\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`SELECT id FROM mz_clusters WHERE name = 'cluster'`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name"}).AddRow("cluster")
		mock.ExpectQuery(`SELECT name FROM mz_clusters WHERE id = 'u1';`).WillReturnRows(ip)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceClusterDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER cluster;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestClusterCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`CREATE CLUSTER cluster REPLICAS ();`, b.Create())
}

func TestClusterDropQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`DROP CLUSTER cluster;`, b.Drop())
}

func TestClusterReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newClusterBuilder("cluster")
	r.Equal(`SELECT id FROM mz_clusters WHERE name = 'cluster';`, b.ReadId())
}

func TestClusterReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := readClusterParams("u1")
	r.Equal(`SELECT name FROM mz_clusters WHERE id = 'u1';`, b)
}
