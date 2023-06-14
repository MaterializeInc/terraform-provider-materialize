package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var readSource = `
SELECT
	mz_sources.id,
	mz_sources.name,
	mz_schemas.name AS schema_name,
	mz_databases.name AS database_name,
	mz_sources.type AS source_type,
	mz_sources.size,
	mz_sources.envelope_type,
	mz_connections.name as connection_name,
	mz_clusters.name as cluster_name
FROM mz_sources
JOIN mz_schemas
	ON mz_sources.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
LEFT JOIN mz_connections
	ON mz_sources.connection_id = mz_connections.id
LEFT JOIN mz_clusters
	ON mz_sources.cluster_id = mz_clusters.id
WHERE mz_sources.id = 'u1';`

func TestResourceSourceUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, inSourcePostgres)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_source")
	d.Set("size", "medium")
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."old_source" SET \(SIZE = 'small'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		ip := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "cluster_name"}).
			AddRow("u1", "source", "schema", "database", "kafka", "small", "JSON", "conn", "cluster")
		mock.ExpectQuery(readSource).WillReturnRows(ip)

		if err := sourceUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "source",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourcePostgres().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
