package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestClusterReplicaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', DISK, AVAILABILITY ZONE = 'us-east-1', INTROSPECTION INTERVAL = '1s', INTROSPECTION DEBUGGING = TRUE, IDLE ARRANGEMENT MERGE EFFORT = 1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "replica", ClusterName: "cluster"}
		b := NewClusterReplicaBuilder(db, o)
		b.Size("xsmall")
		b.Disk(true)
		b.AvailabilityZone("us-east-1")
		b.IntrospectionInterval("1s")
		b.IntrospectionDebugging()
		b.IdleArrangementMergeEffort(1)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterReplicaDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER REPLICA "cluster"."replica";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "replica", ClusterName: "cluster"}
		if err := NewClusterReplicaBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
