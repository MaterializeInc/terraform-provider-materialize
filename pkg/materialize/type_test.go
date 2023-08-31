package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestTypeCreateList(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TYPE "database"."schema"."type" AS LIST \(ELEMENT TYPE = int4\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "type", SchemaName: "schema", DatabaseName: "database"}
		b := NewTypeBuilder(db, o)
		b.ListProperties([]ListProperties{
			{
				ElementType: "int4",
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTypeCreateMap(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TYPE "database"."schema"."type" AS MAP \(KEY TYPE text, VALUE TYPE = int\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "type", SchemaName: "schema", DatabaseName: "database"}
		b := NewTypeBuilder(db, o)
		b.MapProperties([]MapProperties{
			{
				KeyType:   "text",
				ValueType: "int",
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTypeDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TYPE "database"."schema"."type";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "type", SchemaName: "schema", DatabaseName: "database"}
		if err := NewTypeBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
