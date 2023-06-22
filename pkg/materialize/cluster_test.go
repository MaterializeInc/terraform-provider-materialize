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

func TestClusterDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "cluster"}
		if err := NewClusterBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
