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

var inSinkIceberg = map[string]interface{}{
	"name":          "iceberg_sink",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "my_cluster",
	"from": []interface{}{
		map[string]interface{}{
			"name":          "my_view",
			"schema_name":   "public",
			"database_name": "database",
		},
	},
	"iceberg_catalog_connection": []interface{}{
		map[string]interface{}{
			"name":          "iceberg_catalog",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"namespace": "my_namespace",
	"table":     "my_table",
	"aws_connection": []interface{}{
		map[string]interface{}{
			"name":          "aws_conn",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"key":              []interface{}{"id"},
	"key_not_enforced": false,
	"commit_interval":  "10s",
}

func TestResourceSinkIcebergCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."public"."my_view" INTO ICEBERG CATALOG CONNECTION "materialize"."public"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "materialize"."public"."aws_conn" KEY \(id\) MODE UPSERT WITH \(COMMIT INTERVAL = '10s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'iceberg_sink'`
		testhelpers.MockSinkScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkIcebergCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkIcebergCreateWithKeyNotEnforced(t *testing.T) {
	r := require.New(t)
	inSinkIcebergWithNotEnforced := map[string]interface{}{
		"name":          "iceberg_sink",
		"schema_name":   "schema",
		"database_name": "database",
		"cluster_name":  "my_cluster",
		"from": []interface{}{
			map[string]interface{}{
				"name":          "my_view",
				"schema_name":   "public",
				"database_name": "database",
			},
		},
		"iceberg_catalog_connection": []interface{}{
			map[string]interface{}{
				"name":          "iceberg_catalog",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"namespace": "my_namespace",
		"table":     "my_table",
		"aws_connection": []interface{}{
			map[string]interface{}{
				"name":          "aws_conn",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"key":              []interface{}{"id"},
		"key_not_enforced": true,
		"commit_interval":  "30s",
	}
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIcebergWithNotEnforced)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."public"."my_view" INTO ICEBERG CATALOG CONNECTION "materialize"."public"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "materialize"."public"."aws_conn" KEY \(id\) NOT ENFORCED MODE UPSERT WITH \(COMMIT INTERVAL = '30s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'iceberg_sink'`
		testhelpers.MockSinkScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkIcebergCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkIcebergRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	r.NotNil(d)

	// Set id before read
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}
