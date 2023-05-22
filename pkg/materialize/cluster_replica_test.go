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
			`CREATE CLUSTER REPLICA "cluster"."replica" SIZE = 'xsmall', AVAILABILITY ZONE = 'us-east-1', INTROSPECTION INTERVAL = '1s', INTROSPECTION DEBUGGING = TRUE, IDLE ARRANGEMENT MERGE EFFORT = 1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewClusterReplicaBuilder(db, "replica", "cluster")
		b.Size("xsmall")
		b.AvailabilityZone("us-east-1")
		b.IntrospectionInterval("1s")
		b.IntrospectionDebugging()
		b.IdleArrangementMergeEffort(1)

		b.Create()
	})
}

func TestClusterReplicaDropQuery(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP CLUSTER REPLICA "cluster"."replica";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		NewClusterReplicaBuilder(db, "replica", "cluster").Drop()
	})
}
