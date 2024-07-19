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
	"envelope": []interface{}{
		map[string]interface{}{
			"upsert": true,
			"upsert_options": []interface{}{
				map[string]interface{}{
					"value_decoding_errors": []interface{}{
						map[string]interface{}{
							"inline": []interface{}{
								map[string]interface{}{
									"enabled": true,
									"alias":   "my_error_col",
								},
							},
						},
					},
				},
			},
		},
	},
	"start_offset":    []interface{}{1, 2, 3},
	"start_timestamp": -1000,
}

var inSourceKafkaText = map[string]interface{}{
	"name":             "source_text",
	"schema_name":      "schema",
	"database_name":    "database",
	"cluster_name":     "cluster",
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic_text",
	"key_format": []interface{}{
		map[string]interface{}{
			"json": true,
		},
	},
	"value_format": []interface{}{
		map[string]interface{}{
			"text": true,
		},
	},
}

var inSourceKafkaJSON = map[string]interface{}{
	"name":             "source_json",
	"schema_name":      "schema",
	"database_name":    "database",
	"cluster_name":     "cluster",
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic_json",
	"key_format": []interface{}{
		map[string]interface{}{
			"bytes": true,
		},
	},
	"value_format": []interface{}{
		map[string]interface{}{
			"json": true,
		},
	},
}

var inSourceKafkaBytes = map[string]interface{}{
	"name":             "source_bytes",
	"schema_name":      "schema",
	"database_name":    "database",
	"cluster_name":     "cluster",
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic_bytes",
	"key_format": []interface{}{
		map[string]interface{}{
			"text": true,
		},
	},
	"value_format": []interface{}{
		map[string]interface{}{
			"bytes": true,
		},
	},
}

var inSourceKafkaCSV = map[string]interface{}{
	"name":             "source_csv",
	"schema_name":      "schema",
	"database_name":    "database",
	"cluster_name":     "cluster",
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic_csv",
	"key_format": []interface{}{
		map[string]interface{}{
			"json": true,
		},
	},
	"value_format": []interface{}{
		map[string]interface{}{
			"csv": []interface{}{
				map[string]interface{}{
					"delimited_by": ",",
					"header":       []interface{}{"column1", "column2", "column3"},
				},
			},
		},
	},
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
			ENVELOPE UPSERT \(VALUE DECODING ERRORS = \(INLINE AS my_error_col\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
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

	testInSourceKafka["envelope"] = []interface{}{
		map[string]interface{}{
			"upsert": true,
		},
	}

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
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
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

	testInSourceKafka["envelope"] = []interface{}{
		map[string]interface{}{
			"debezium": true,
		},
	}

	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, testInSourceKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
			IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic', START TIMESTAMP -1000, START OFFSET \(1,2,3\)\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" VALUE STRATEGY avro_key_fullname
			ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateTextFormat(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafkaText)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source_text"
            IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic_text'\)
            KEY FORMAT JSON
            VALUE FORMAT TEXT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source_text'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateJSONFormat(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafkaJSON)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source_json"
            IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic_json'\)
            KEY FORMAT BYTES
            VALUE FORMAT JSON;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source_json'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateBytesFormat(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafkaBytes)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source_bytes"
            IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic_bytes'\)
            KEY FORMAT TEXT
            VALUE FORMAT BYTES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source_bytes'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateCSVFormat(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceKafka().Schema, inSourceKafkaCSV)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source_csv"
            IN CLUSTER "cluster" FROM KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic_csv'\)
            KEY FORMAT JSON
            VALUE FORMAT CSV WITH HEADER \( column1, column2, column3 \) DELIMITER ',';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source_csv'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE filter_id = 'u1' AND type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
