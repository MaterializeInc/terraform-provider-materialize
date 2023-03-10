package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceDatabaseCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "database",
	}
	d := schema.TestResourceDataRaw(t, Database().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE DATABASE database;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`SELECT id FROM mz_databases WHERE name = 'database'`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name"}).
			AddRow("database")
		mock.ExpectQuery(`SELECT name FROM mz_databases WHERE id = 'u1';`).WillReturnRows(ip)

		if err := databaseCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceDatabaseDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "database",
	}
	d := schema.TestResourceDataRaw(t, Database().Schema, in)
	r.NotNil(d)

	WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP DATABASE database;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := databaseDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestDatabaseReadIdQuery(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`SELECT id FROM mz_databases WHERE name = 'database';`, b.ReadId())
}

func TestDatabaseCreateQuery(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`CREATE DATABASE database;`, b.Create())
}

func TestDatabaseDropQuery(t *testing.T) {
	r := require.New(t)
	b := newDatabaseBuilder("database")
	r.Equal(`DROP DATABASE database;`, b.Drop())
}

func TestDatabaseReadParamsQuery(t *testing.T) {
	r := require.New(t)
	b := readDatabaseParams("u1")
	r.Equal(`SELECT name FROM mz_databases WHERE id = 'u1';`, b)
}
