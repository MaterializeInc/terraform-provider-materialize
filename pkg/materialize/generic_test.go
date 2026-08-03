package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestBuilderExecEmptyStatement(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		b := Builder{db, Cluster}

		if err := b.exec(""); err == nil {
			t.Fatal("Expected an error for an empty statement")
		}
	})
}
