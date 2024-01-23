package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceLoadgen = MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}

func TestSourceLoadgenCounterCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			FROM LOAD GENERATOR COUNTER
			\(TICK INTERVAL '1s', MAX CARDINALITY 8\)
			EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, sourceLoadgen)
		b.LoadGeneratorType("COUNTER")
		b.CounterOptions(CounterOptions{
			TickInterval:   "1s",
			MaxCardinality: 8,
		})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceLoadgenAuctionCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			FROM LOAD GENERATOR AUCTION
			\(TICK INTERVAL '1s', SCALE FACTOR 0.01\)
			FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, sourceLoadgen)
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

func TestSourceLoadgenMarketingCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			FROM LOAD GENERATOR MARKETING
			\(TICK INTERVAL '1s', SCALE FACTOR 0.01\)
			FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, sourceLoadgen)
		b.LoadGeneratorType("MARKETING")
		b.MarketingOptions(MarketingOptions{
			TickInterval: "1s",
			ScaleFactor:  0.01,
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceLoadgenTPCHParamsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			FROM LOAD GENERATOR TPCH
			\(TICK INTERVAL '1s', SCALE FACTOR 0.01\)
			FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceLoadgenBuilder(db, sourceLoadgen)
		b.LoadGeneratorType("TPCH")
		b.TPCHOptions(TPCHOptions{
			TickInterval: "1s",
			ScaleFactor:  0.01,
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
