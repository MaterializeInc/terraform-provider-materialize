package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://github.com/MaterializeInc/materialize/blob/main/test/sqllogictest/managed_cluster.slt
// https://materialize.com/docs/sql/create-cluster/

func TestClusterCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" REPLICAS \(\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" SIZE 'xsmall';`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedReplicationFactorCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" SIZE 'xsmall', REPLICATION FACTOR 3;`).WillReturnResult(sqlmock.NewResult(1, 1))
		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")
		r := 3
		b.ReplicationFactor(&r)
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedSizeDiskCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" SIZE 'xsmall', DISK;`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")
		b.Disk(true)
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedAllCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`
			CREATE CLUSTER "cluster"
			SIZE 'xsmall',
			REPLICATION FACTOR 2,
			AVAILABILITY ZONES = \['us-east-1'\],
			INTROSPECTION INTERVAL = '1s',
			INTROSPECTION DEBUGGING = TRUE;
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")
		r := 2
		b.ReplicationFactor(&r)
		b.AvailabilityZones([]string{"us-east-1"})
		b.IntrospectionInterval("1s")
		b.IntrospectionDebugging()
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterWithSchedulingCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `CREATE CLUSTER "cluster" SIZE 'xsmall', SCHEDULE = ON REFRESH \(REHYDRATION TIME ESTIMATE = '2 hours'\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")

		b.Scheduling([]interface{}{
			map[string]interface{}{
				"on_refresh": []interface{}{
					map[string]interface{}{
						"enabled":                   true,
						"rehydration_time_estimate": "2 hours",
					},
				},
			},
		})

		if err := b.Create(); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}
