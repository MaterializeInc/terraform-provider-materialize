package materialize

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSourceLoadgenCreateCounterParams(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR COUNTER \(TICK INTERVAL '1s', MAX CARDINALITY 8\) WITH \(SIZE = 'xsmall'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, "source", "schema", "database")
		b.Size("xsmall")
		b.LoadGeneratorType("COUNTER")
		b.CounterOptions(CounterOptions{
			TickInterval:   "1s",
			MaxCardinality: 8,
		})
		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceLoadgenCreateAuction(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR AUCTION \(TICK INTERVAL '1s', SCALE FACTOR 0.01\) FOR ALL TABLES WITH \(SIZE = 'xsmall'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, "source", "schema", "database")
		b.Size("xsmall")
		b.LoadGeneratorType("AUCTION")
		b.AuctionOptions(AuctionOptions{
			TickInterval: "1s",
			ScaleFactor:  0.01,
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceLoadgenCreateTPCH(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM LOAD GENERATOR TPCH \(TICK INTERVAL '1s', SCALE FACTOR 0.01\) FOR TABLES \(schema1.table_1 AS s1_table_1, schema2.table_1 AS s2_table_1\) WITH \(SIZE = 'xsmall'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, "source", "schema", "database")
		b.Size("xsmall")
		b.LoadGeneratorType("TPCH")
		b.TPCHOptions(TPCHOptions{
			TickInterval: "1s",
			ScaleFactor:  0.01,
			Table: []Table{
				{
					Name:  "schema1.table_1",
					Alias: "s1_table_1",
				},
				{
					Name:  "schema2.table_1",
					Alias: "s2_table_1",
				},
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
