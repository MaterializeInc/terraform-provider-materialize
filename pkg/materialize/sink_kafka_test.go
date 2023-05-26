package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSinkKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink" FROM "database"."schema"."table" INTO KAFKA CONNECTION "database"."schema"."kafka_connection" KEY (key_1, key_2) (TOPIC 'test_avro_topic') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."public"."csr_connection" ENVELOPE UPSERT WITH ( SIZE = 'xsmall' SNAPSHOT = false);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSinkKafkaBuilder(db, "sink", "schema", "database")
		b.Size("xsmall")
		b.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("test_avro_topic")
		b.Key([]string{"key_1", "key_2"})
		b.Format(SinkFormatSpecStruct{Avro: &SinkAvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "public"}}})
		b.Envelope(KafkaSinkEnvelopeStruct{Upsert: true})
		b.Snapshot(false)

		b.Create()
	})
}
