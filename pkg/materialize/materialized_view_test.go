package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestMaterializedViewCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE MATERIALIZED VIEW "database"."schema"."materialized_view" IN CLUSTER "cluster" WITH \(ASSERT NOT NULL "column_1", ASSERT NOT NULL "column_2"\) AS SELECT 1 FROM t1;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "materialized_view", SchemaName: "schema", DatabaseName: "database"}
		b := NewMaterializedViewBuilder(db, o)
		b.ClusterName("cluster")
		b.NotNullAssertions([]string{"column_1", "column_2"})
		b.SelectStmt("SELECT 1 FROM t1")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMaterializedViewDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP MATERIALIZED VIEW "database"."schema"."materialized_view";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "materialized_view", SchemaName: "schema", DatabaseName: "database"}
		if err := NewMaterializedViewBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMaterializedAlterCluster(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER MATERIALIZED VIEW "database"."schema"."materialized_view" SET CLUSTER "new_cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "materialized_view", SchemaName: "schema", DatabaseName: "database"}
		if err := NewMaterializedViewBuilder(db, o).AlterCluster("new_cluster"); err != nil {
			t.Fatal(err)
		}
	})
}
