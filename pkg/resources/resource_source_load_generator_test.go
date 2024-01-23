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

		// Query Subsources
		ps := `WHERE mz_object_dependencies.object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceLoadgenCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
