package materialize

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableCreateQuery(t *testing.T) {
	r := require.New(t)
	b := NewTableBuilder("table", "schema", "database")
	b.Column([]TableColumn{
		{
			ColName: "column_1",
			ColType: "int",
		},
		{
			ColName: "column_2",
			ColType: "text",
			NotNull: true,
		},
	})
	r.Equal(`CREATE TABLE "database"."schema"."table" (column_1 int, column_2 text NOT NULL);`, b.Create())
}

func TestTableRenameQuery(t *testing.T) {
	r := require.New(t)
	b := NewTableBuilder("table", "schema", "database")
	r.Equal(`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`, b.Rename("new_table"))
}

func TestTableDropQuery(t *testing.T) {
	r := require.New(t)
	b := NewTableBuilder("table", "schema", "database")
	r.Equal(`DROP TABLE "database"."schema"."table";`, b.Drop())
}
