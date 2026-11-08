package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestDatabaseCreate(t *testing.T) {
	withMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`CREATE DATABASE "database";`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "database"}
		if err := NewDatabaseBuilder(db, o).Create(); err != nil {
			t.Fatal(err)
		}
	})
}
