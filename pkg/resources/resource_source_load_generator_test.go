package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inSourceLoadgen = map[string]interface{}{
	"name":                "source",
	"schema_name":         "schema",
	"database_name":       "database",
	"cluster_name":        "cluster",
	"expose_progress":     []interface{}{map[string]interface{}{"name": "progress"}},
	"load_generator_type": "TPCH",
	"tpch_options": []interface{}{map[string]interface{}{
		"tick_interval": "1s",
		"scale_factor":  0.5,
	}},
}

func TestResourceSourceLoadgenCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, inSourceLoadgen)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster"
			FROM LOAD GENERATOR TPCH
			\(TICK INTERVAL '1s', SCALE FACTOR 0.50\)
			FOR ALL TABLES
			EXPOSE PROGRESS AS "materialize"."public"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		if err := sourceLoadgenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceLoadgenKeyValueCreate(t *testing.T) {
	r := require.New(t)
	inSourceLoadgenKeyValue := map[string]interface{}{
		"name":                "source",
		"schema_name":         "schema",
		"database_name":       "database",
		"cluster_name":        "cluster",
		"expose_progress":     []interface{}{map[string]interface{}{"name": "progress"}},
		"load_generator_type": "KEY VALUE",
		"key_value_options": []interface{}{map[string]interface{}{
			"keys":                   200,
			"snapshot_rounds":        5,
			"transactional_snapshot": true,
			"value_size":             256,
			"tick_interval":          "2s",
			"seed":                   42,
			"partitions":             10,
			"batch_size":             20,
		}},
	}

	d := schema.TestResourceDataRaw(t, SourceLoadgen().Schema, inSourceLoadgenKeyValue)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster"
			FROM LOAD GENERATOR KEY VALUE
			\(KEYS 200, SNAPSHOT ROUNDS 5, TRANSACTIONAL SNAPSHOT true, VALUE SIZE 256, TICK INTERVAL '2s', SEED 42, PARTITIONS 10, BATCH SIZE 20\)
			EXPOSE PROGRESS AS "materialize"."public"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		if err := sourceLoadgenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
