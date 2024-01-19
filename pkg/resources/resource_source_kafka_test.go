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

var inSourceKafka = map[string]interface{}{
	"name":                    "source",
	"schema_name":             "schema",
	"database_name":           "database",
	"cluster_name":            "cluster",
	"item_name":               "item",
	"kafka_connection":        []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":                   "topic",
	"include_key":             true,
	"include_key_alias":       "key",
	"include_headers":         true,
	"include_headers_alias":   "headers",
	"include_partition":       true,
	"include_partition_alias": "partition",
	"include_offset":          true,
	"include_offset_alias":    "offset",
	"include_timestamp":       true,
	"include_timestamp_alias": "timestamp",
	"format": []interface{}{
		map[string]interface{}{
			"avro": []interface{}{
				map[string]interface{}{
					"value_strategy": "avro_key_fullname",
					"schema_registry_connection": []interface{}{
						map[string]interface{}{
							"name":          "csr_conn",
							"database_name": "database",
							"schema_name":   "schema",
						},
					},
				},
			},
		},
	},
	"envelope":        []interface{}{map[string]interface{}{"upsert": true}},
	"start_offset":    []interface{}{1, 2, 3},
	"start_timestamp": -1000,
}

func TestResourceSourceKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic', START TIMESTAMP -1000, START OFFSET \(1,2,3\)\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" VALUE STRATEGY avro_key_fullname
			INCLUDE KEY AS key,
			HEADERS AS headers,
			PARTITION AS partition,
			OFFSET AS offset,
			TIMESTAMP AS timestamp
			ENVELOPE UPSERT;`,
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

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateIncludeTrueNoAlias(t *testing.T) {
	r := require.New(t)

	testInSourceKafka := inSourceKafka
	testInSourceKafka["include_key"] = true
	delete(testInSourceKafka, "include_key_alias")
	testInSourceKafka["include_headers"] = true
	delete(testInSourceKafka, "include_headers_alias")
	testInSourceKafka["include_partition"] = true
	delete(testInSourceKafka, "include_partition_alias")
	testInSourceKafka["include_offset"] = true
	delete(testInSourceKafka, "include_offset_alias")
	testInSourceKafka["include_timestamp"] = true
	delete(testInSourceKafka, "include_timestamp_alias")

	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, testInSourceKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic', START TIMESTAMP -1000, START OFFSET \(1,2,3\)\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" VALUE STRATEGY avro_key_fullname
			INCLUDE KEY,
			HEADERS,
			PARTITION,
			OFFSET,
			TIMESTAMP
			ENVELOPE UPSERT;`,
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

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateIncludeFalseWithAlias(t *testing.T) {
	r := require.New(t)

	testInSourceKafka := inSourceKafka
	testInSourceKafka["include_key"] = false
	testInSourceKafka["include_headers"] = false
	testInSourceKafka["include_partition"] = false
	testInSourceKafka["include_offset"] = false
	testInSourceKafka["include_timestamp"] = false

	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, testInSourceKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic', START TIMESTAMP -1000, START OFFSET \(1,2,3\)\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" VALUE STRATEGY avro_key_fullname
			ENVELOPE UPSERT
			WITH \(SIZE = 'small'\);`,
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

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
