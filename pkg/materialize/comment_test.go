package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestCommentObject(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`COMMENT ON TABLE "database"."schema"."table" IS 'my comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{ObjectType: "TABLE", Name: "table", DatabaseName: "database", SchemaName: "schema"}
		c := "my comment"
		if err := NewCommentBuilder(db, o).Object(c); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCommentColumn(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`COMMENT ON COLUMN "database"."schema"."table"."column" IS 'my comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{ObjectType: "TABLE", Name: "table", DatabaseName: "database", SchemaName: "schema"}
		c := "my comment"
		if err := NewCommentBuilder(db, o).Column("column", c); err != nil {
			t.Fatal(err)
		}
	})
}
