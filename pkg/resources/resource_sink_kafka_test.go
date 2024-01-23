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

var inSinkKafka = map[string]interface{}{
	"name":          "sink",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"from": []interface{}{
		map[string]interface{}{
			"name":          "item",
			"schema_name":   "public",
			"database_name": "database",
		},
	},
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic",
	"compression_type": "gzip",
	"key":              []interface{}{"key_1", "key_2"},
	"key_not_enforced": true,
	"format": []interface{}{
		map[string]interface{}{
			"avro": []interface{}{
				map[string]interface{}{
					"avro_key_fullname":   "avro_key_fullname",
					"avro_value_fullname": "avro_value_fullname",
					"schema_registry_connection": []interface{}{
						map[string]interface{}{
							"name":          "csr_conn",
							"database_name": "database",
							"schema_name":   "schema",
						},
					},
					"avro_doc_type": []interface{}{
						map[string]interface{}{
							"object": []interface{}{
								map[string]interface{}{
									"name":          "item",
									"schema_name":   "public",
									"database_name": "database",
								},
							},
							"doc": "top-level comment",
						},
					},
					"avro_doc_column": []interface{}{
						map[string]interface{}{
							"object": []interface{}{
								map[string]interface{}{
									"name":          "item",
									"schema_name":   "public",
									"database_name": "database",
								},
							},
							"column": "c1",
							"doc":    "comment on column only in key schema",
							"key":    true,
						},
						map[string]interface{}{
							"object": []interface{}{
								map[string]interface{}{
									"name":          "item",
									"schema_name":   "public",
									"database_name": "database",
								},
							},
							"column": "c1",
							"doc":    "comment on column only in value schema",
							"value":  true,
						},
					},
				},
			},
		},
	},
	"envelope": []interface{}{map[string]interface{}{"upsert": true}},
	"snapshot": false,
}

func TestResourceSinkKafkaCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkKafka().Schema, inSinkKafka)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			IN CLUSTER "cluster" FROM "database"."public"."item"
			INTO KAFKA CONNECTION "materialize"."public"."kafka_conn"
			\(TOPIC 'topic', COMPRESSION TYPE = gzip\) KEY \(key_1, key_2\)
			NOT ENFORCED FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn"
			\(AVRO KEY FULLNAME 'avro_key_fullname' AVRO VALUE FULLNAME 'avro_value_fullname',
			DOC ON TYPE "database"."public"."item" = 'top-level comment',
			KEY DOC ON COLUMN "database"."public"."item"."c1" = 'comment on column only in key schema',
			VALUE DOC ON COLUMN "database"."public"."item"."c1" = 'comment on column only in value schema'\)
			ENVELOPE UPSERT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'sink'`
		testhelpers.MockSinkScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkKafkaCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
