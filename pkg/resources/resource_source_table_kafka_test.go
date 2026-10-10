package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

var inSourceTableKafka = map[string]interface{}{
	"name":          "table",
	"schema_name":   "schema",
	"database_name": "database",
	"source": []interface{}{
		map[string]interface{}{
			"name":          "kafka_source",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"topic":                   "topic",
	"include_key":             true,
	"include_key_alias":       "message_key",
	"include_headers":         true,
	"include_headers_alias":   "message_headers",
	"include_partition":       true,
	"include_partition_alias": "message_partition",
	"format": []interface{}{
		map[string]interface{}{
			"json": true,
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
									"alias":   "decoding_error",
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestResourceSourceTableKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT JSON
            INCLUDE KEY AS "message_key", HEADERS AS "message_headers", PARTITION AS "message_partition"
            ENVELOPE UPSERT \(VALUE DECODING ERRORS = \(INLINE AS "decoding_error"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafka)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("table", d.Get("name").(string))
		r.Equal("schema", d.Get("schema_name").(string))
		r.Equal("database", d.Get("database_name").(string))
	})
}

func TestResourceSourceTableKafkaUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafka)
	d.SetId("u1")
	d.Set("name", "old_table")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."" RENAME TO "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafka)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithAvroFormat(t *testing.T) {
	r := require.New(t)
	inSourceTableKafkaAvro := map[string]interface{}{
		"name":          "table_avro",
		"schema_name":   "schema",
		"database_name": "database",
		"source": []interface{}{
			map[string]interface{}{
				"name":          "kafka_source",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"topic": "topic",
		"format": []interface{}{
			map[string]interface{}{
				"avro": []interface{}{
					map[string]interface{}{
						"schema_registry_connection": []interface{}{
							map[string]interface{}{
								"name":          "sr_conn",
								"schema_name":   "public",
								"database_name": "materialize",
							},
						},
					},
				},
			},
		},
		"envelope": []interface{}{
			map[string]interface{}{
				"debezium": true,
			},
		},
	}
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafkaAvro)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table_avro"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."sr_conn"
            ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table_avro'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateIncludeTrueNoAlias(t *testing.T) {
	r := require.New(t)

	testInSourceTableKafka := inSourceTableKafka
	testInSourceTableKafka["include_key"] = true
	delete(testInSourceTableKafka, "include_key_alias")
	testInSourceTableKafka["include_headers"] = true
	delete(testInSourceTableKafka, "include_headers_alias")
	testInSourceTableKafka["include_partition"] = true
	delete(testInSourceTableKafka, "include_partition_alias")
	testInSourceTableKafka["include_offset"] = true
	testInSourceTableKafka["include_timestamp"] = true

	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, testInSourceTableKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT JSON
            INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP
            ENVELOPE UPSERT \(VALUE DECODING ERRORS = \(INLINE AS "decoding_error"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateIncludeFalseWithAlias(t *testing.T) {
	r := require.New(t)

	testInSourceTableKafka := inSourceTableKafka
	testInSourceTableKafka["include_key"] = false
	testInSourceTableKafka["include_headers"] = false
	testInSourceTableKafka["include_partition"] = false
	testInSourceTableKafka["include_offset"] = false
	testInSourceTableKafka["include_timestamp"] = false

	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, testInSourceTableKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT JSON
            ENVELOPE UPSERT \(VALUE DECODING ERRORS = \(INLINE AS "decoding_error"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithCSVFormat(t *testing.T) {
	r := require.New(t)
	inSourceTableKafkaCSV := map[string]interface{}{
		"name":          "table_csv",
		"schema_name":   "schema",
		"database_name": "database",
		"source": []interface{}{
			map[string]interface{}{
				"name":          "kafka_source",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"topic": "topic",
		"format": []interface{}{
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
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafkaCSV)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table_csv"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT CSV WITH HEADER \( column1, column2, column3 \) DELIMITER ',';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table_csv'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithKeyAndValueFormat(t *testing.T) {
	r := require.New(t)
	inSourceTableKafkaKeyValue := map[string]interface{}{
		"name":          "table_key_value",
		"schema_name":   "schema",
		"database_name": "database",
		"source": []interface{}{
			map[string]interface{}{
				"name":          "kafka_source",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"topic": "topic",
		"key_format": []interface{}{
			map[string]interface{}{
				"json": true,
			},
		},
		"value_format": []interface{}{
			map[string]interface{}{
				"avro": []interface{}{
					map[string]interface{}{
						"schema_registry_connection": []interface{}{
							map[string]interface{}{
								"name":          "sr_conn",
								"schema_name":   "public",
								"database_name": "materialize",
							},
						},
					},
				},
			},
		},
	}
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafkaKeyValue)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table_key_value"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            KEY FORMAT JSON
            VALUE FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."sr_conn";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table_key_value'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithProtobufFormat(t *testing.T) {
	r := require.New(t)
	inSourceTableKafkaProtobuf := map[string]interface{}{
		"name":          "table_protobuf",
		"schema_name":   "schema",
		"database_name": "database",
		"source": []interface{}{
			map[string]interface{}{
				"name":          "kafka_source",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"topic": "topic",
		"format": []interface{}{
			map[string]interface{}{
				"protobuf": []interface{}{
					map[string]interface{}{
						"schema_registry_connection": []interface{}{
							map[string]interface{}{
								"name":          "sr_conn",
								"schema_name":   "public",
								"database_name": "materialize",
							},
						},
						"message": "MyMessage",
					},
				},
			},
		},
		"envelope": []interface{}{
			map[string]interface{}{
				"none": true,
			},
		},
	}
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafkaProtobuf)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table_protobuf"
            FROM SOURCE "materialize"."public"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT PROTOBUF MESSAGE 'MyMessage' USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."sr_conn"
            ENVELOPE NONE;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table_protobuf'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithNoTopic(t *testing.T) {
	r := require.New(t)
	inSourceTableKafkaNoTopic := map[string]interface{}{
		"name":          "no_topic",
		"schema_name":   "schema",
		"database_name": "database",
		"source": []interface{}{
			map[string]interface{}{
				"name":          "kafka_source",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"format": []interface{}{
			map[string]interface{}{
				"json": true,
			},
		},
		"envelope": []interface{}{
			map[string]interface{}{
				"none": true,
			},
		},
	}
	d := schema.TestResourceDataRaw(t, SourceTableKafka().Schema, inSourceTableKafkaNoTopic)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."no_topic"
            FROM SOURCE "materialize"."public"."kafka_source"
            FORMAT JSON
            ENVELOPE NONE;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'no_topic'`
		testhelpers.MockSourceTableKafkaScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableKafkaScan(mock, pp)

		if err := sourceTableKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Materialize takes either FORMAT or KEY FORMAT ... VALUE FORMAT on CREATE
// TABLE ... FROM SOURCE, and the key and value formats only together. The plan
// rejects the other combinations instead of failing at apply.
func TestResourceSourceTableKafkaFormatCombinations(t *testing.T) {
	text := []interface{}{map[string]interface{}{"text": true}}
	validate := func(formats map[string]interface{}) error {
		cfg := map[string]interface{}{
			"name":   "table",
			"source": []interface{}{map[string]interface{}{"name": "kafka_source"}},
		}
		for k, v := range formats {
			cfg[k] = v
		}
		var errs []string
		for _, d := range SourceTableKafka().Validate(terraform.NewResourceConfigRaw(cfg)) {
			if d.Severity == diag.Error {
				errs = append(errs, d.Summary+": "+d.Detail)
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	cases := []struct {
		name    string
		formats map[string]interface{}
		wantErr string
	}{
		{"no format", nil, ""},
		{"format only", map[string]interface{}{"format": text}, ""},
		{"key and value format", map[string]interface{}{"key_format": text, "value_format": text}, ""},
		{"format with key_format", map[string]interface{}{"format": text, "key_format": text}, "conflicts with"},
		{"format with value_format", map[string]interface{}{"format": text, "value_format": text}, "conflicts with"},
		{"format with both", map[string]interface{}{"format": text, "key_format": text, "value_format": text}, "conflicts with"},
		{"key_format only", map[string]interface{}{"key_format": text}, "all of `key_format,value_format` must be specified"},
		{"value_format only", map[string]interface{}{"value_format": text}, "all of `key_format,value_format` must be specified"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validate(c.formats)
			if c.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), c.wantErr)
		})
	}
}
