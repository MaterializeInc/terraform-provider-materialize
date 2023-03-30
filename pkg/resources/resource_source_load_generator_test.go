package resources

import (
	"context"
	"terraform-materialize/pkg/testhelpers"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceSourceLoadgenCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":                "source",
		"schema_name":         "schema",
		"database_name":       "database",
		"cluster_name":        "cluster",
		"size":                "small",
		"load_generator_type": "TPCH",
		"tick_interval":       "1s",
		"scale_factor":        0.5,
		"max_cardinality":     true,
		"table":               []interface{}{map[string]interface{}{"name": "name", "alias": "alias"}},
	}
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "cluster" FROM LOAD GENERATOR TPCH \(TICK INTERVAL '1s', SCALE FACTOR 0.50, MAX CARDINALITY\) FOR TABLES \(name AS alias\) WITH \(SIZE = 'small'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(`
			SELECT mz_sources.id
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sources.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sources.cluster_id = mz_clusters.id
			WHERE mz_sources.name = 'source'
			AND mz_schemas.name = 'schema'
			AND mz_databases.name = 'database';
		`).WillReturnRows(ir)

		// Query Params
		ip := sqlmock.NewRows([]string{"name", "schema", "database", "source_type", "size", "connection_name", "cluster_name"}).
			AddRow("conn", "schema", "database", "source_type", "small", "conn", "cluster")
		mock.ExpectQuery(`
			SELECT
				mz_sources.name,
				mz_schemas.name,
				mz_databases.name,
				mz_sources.type,
				mz_sources.size,
				mz_connections.name as connection_name,
				mz_clusters.name as cluster_name
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			LEFT JOIN mz_connections
				ON mz_sources.connection_id = mz_connections.id
			JOIN mz_clusters
				ON mz_sources.cluster_id = mz_clusters.id
			WHERE mz_sources.id = 'u1';`).WillReturnRows(ip)

		if err := sourceLoadgenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

func TestResourceSourceLoadgenDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":          "source",
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceLoadgenDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
