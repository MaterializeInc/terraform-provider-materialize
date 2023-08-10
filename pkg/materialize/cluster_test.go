package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestClusterCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" REPLICAS \(\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" SIZE 'xsmall', REPLICATION FACTOR 2 AVAILABILITY ZONE = \['us-east-1'\], INTROSPECTION INTERVAL = '1s', INTROSPECTION DEBUGGING = TRUE, IDLE ARRANGEMENT MERGE EFFORT = 1;`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.ReplicationFactor(2)
		b.Size("xsmall")
		b.AvailabilityZones([]string{"us-east-1"})
		b.IntrospectionInterval("1s")
		b.IntrospectionDebugging()
		b.IdleArrangementMergeEffort(1)
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
