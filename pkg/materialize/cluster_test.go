package materialize

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
)

// https://github.com/MaterializeInc/materialize/blob/main/test/sqllogictest/managed_cluster.slt
// https://materialize.com/docs/sql/create-cluster/

func TestClusterCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" \(REPLICAS \(\)\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClusterManagedCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE CLUSTER "cluster" \(SIZE 'xsmall'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

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
		mock.ExpectExec(`CREATE CLUSTER "cluster" \(SIZE 'xsmall', REPLICATION FACTOR 3\);`).WillReturnResult(sqlmock.NewResult(1, 1))
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
		mock.ExpectExec(`CREATE CLUSTER "cluster" \(SIZE 'xsmall', DISK\);`).WillReturnResult(sqlmock.NewResult(1, 1))

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
			\(SIZE 'xsmall',
			REPLICATION FACTOR 2,
			AVAILABILITY ZONES = \['us-east-1'\],
			INTROSPECTION INTERVAL = '1s',
			INTROSPECTION DEBUGGING = TRUE\);
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
		expectedSQL := `CREATE CLUSTER "cluster" \(SIZE 'xsmall', SCHEDULE = ON REFRESH \(HYDRATION TIME ESTIMATE = '2 hours'\)\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")

		b.Scheduling([]interface{}{
			map[string]interface{}{
				"on_refresh": []interface{}{
					map[string]interface{}{
						"enabled":                 true,
						"hydration_time_estimate": "2 hours",
					},
				},
			},
		})

		if err := b.Create(); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterWithAutoScalingCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `CREATE CLUSTER "cluster" \(SIZE 'xsmall', AUTO SCALING STRATEGY = \(ON HYDRATION \(HYDRATION SIZE = '800cc', LINGER DURATION = '15s'\)\)\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")

		b.AutoScalingStrategy([]interface{}{
			map[string]interface{}{
				"on_hydration": []interface{}{
					map[string]interface{}{
						"hydration_size":  "800cc",
						"linger_duration": "15s",
					},
				},
			},
		})

		if err := b.Create(); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterWithAutoScalingNoLingerCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `CREATE CLUSTER "cluster" \(SIZE 'xsmall', AUTO SCALING STRATEGY = \(ON HYDRATION \(HYDRATION SIZE = '800cc'\)\)\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.Size("xsmall")

		b.AutoScalingStrategy([]interface{}{
			map[string]interface{}{
				"on_hydration": []interface{}{
					map[string]interface{}{
						"hydration_size": "800cc",
					},
				},
			},
		})

		if err := b.Create(); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterAutoScalingUpdate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `ALTER CLUSTER "cluster" SET \(AUTO SCALING STRATEGY = \(ON HYDRATION \(HYDRATION SIZE = '800cc'\)\)\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.SetAutoScalingStrategy([]interface{}{
			map[string]interface{}{
				"on_hydration": []interface{}{
					map[string]interface{}{
						"hydration_size": "800cc",
					},
				},
			},
		})

		if err := b.AlterCluster(ReconfigurationOptions{}); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterAutoScalingReset(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CLUSTER "cluster" RESET \(AUTO SCALING STRATEGY\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)

		if err := b.AlterClusterResetAutoScaling(); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterUpdate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `ALTER CLUSTER "cluster" SET \(SIZE 'xsmall', REPLICATION FACTOR 2\);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.SetSize("xsmall")
		b.SetReplicationFactor(2)

		if err := b.AlterCluster(ReconfigurationOptions{}); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestClusterUpdateWithWaitUntilReady(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedSQL := `ALTER CLUSTER "cluster" SET \(SIZE 'xsmall', REPLICATION FACTOR 2\) WITH \( WAIT UNTIL READY \( TIMEOUT '10s', ON TIMEOUT 'COMMIT' \) \);`
		mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "cluster"}
		b := NewClusterBuilder(db, o)
		b.SetSize("xsmall")
		b.SetReplicationFactor(2)

		if err := b.AlterCluster(ReconfigurationOptions{enabled: true, timeout: "10s", on_timeout: "COMMIT"}); err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestScanClusterAutoScalingStrategyMissingView(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`FROM mz_internal.mz_cluster_auto_scaling_strategies`).
			WillReturnError(&pgconn.PgError{Code: "42P01", Message: "unknown catalog item"})

		s, err := ScanClusterAutoScalingStrategy(db, "u1")
		if err != nil {
			t.Fatalf("Expected a missing view to degrade gracefully, got %v", err)
		}
		if s.HydrationSize.Valid {
			t.Fatal("Expected an empty strategy")
		}
	})
}

func TestScanClusterAutoScalingStrategyQueryError(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`FROM mz_internal.mz_cluster_auto_scaling_strategies`).
			WillReturnError(errors.New("connection reset by peer"))

		if _, err := ScanClusterAutoScalingStrategy(db, "u1"); err == nil {
			t.Fatal("Expected a transient failure to be returned, not swallowed")
		}
	})
}

func TestScanClusterPendingReconfigurationMissingView(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`FROM mz_internal.mz_cluster_reconfigurations`).
			WillReturnError(&pgconn.PgError{Code: "42P01", Message: "unknown catalog item"})

		_, inFlight, err := ScanClusterPendingReconfiguration(db, "u1")
		if err != nil {
			t.Fatalf("Expected a missing view to degrade gracefully, got %v", err)
		}
		if inFlight {
			t.Fatal("Expected no in-flight reconfiguration")
		}
	})
}

func TestScanClusterPendingReconfigurationQueryError(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(`FROM mz_internal.mz_cluster_reconfigurations`).
			WillReturnError(errors.New("connection reset by peer"))

		if _, _, err := ScanClusterPendingReconfiguration(db, "u1"); err == nil {
			t.Fatal("Expected a transient failure to be returned, not swallowed")
		}
	})
}
