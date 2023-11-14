package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inSinkKafka = map[string]interface{}{
	"name":          "sink",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"size":          "small",
	"from": []interface{}{
		map[string]interface{}{
			"name":          "item",
			"schema_name":   "public",
			"database_name": "database",
		},
	},
	"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
	"topic":            "topic",
	"key":              []interface{}{"key_1", "key_2"},
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

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink" IN CLUSTER "cluster" FROM "database"."public"."item" INTO KAFKA CONNECTION "materialize"."public"."kafka_conn" \(TOPIC 'topic'\) KEY \(key_1, key_2\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_conn" WITH \(AVRO KEY FULLNAME 'avro_key_fullname' AVRO VALUE FULLNAME 'avro_value_fullname'\) ENVELOPE UPSERT WITH \( SIZE = 'small' SNAPSHOT = false\);`,
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
