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
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."table" INTO KAFKA CONNECTION "database"."schema"."kafka_connection"
			\(TOPIC 'test_avro_topic'\) KEY \(key_1, key_2\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."public"."csr_connection"
			ENVELOPE UPSERT WITH \( SIZE = 'xsmall' SNAPSHOT = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.Size("xsmall")
		b.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("test_avro_topic")
		b.Key([]string{"key_1", "key_2"})
		b.Format(SinkFormatSpecStruct{Avro: &SinkAvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "public"}}})
		b.Envelope(KafkaSinkEnvelopeStruct{Upsert: true})
		b.Snapshot(false)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaAvroDocsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."table" INTO KAFKA CONNECTION "database"."schema"."kafka_connection" 
			\(TOPIC 'test_avro_topic'\)
			KEY \(key_1, key_2\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."public"."csr_connection" 
			ENVELOPE UPSERT WITH \( SIZE = 'xsmall' SNAPSHOT = false\)
			\(DOC ON TYPE "database"."schema"."table" = 'top-level comment',
			KEY DOC ON COLUMN "database"."schema"."table"."c1" = 'comment on column only in key schema',
			VALUE DOC ON COLUMN TYPE "database"."schema"."table"."c1" = 'comment on column only in value schema',
			KEY DOC ON COLUMN "database"."schema"."table"."c2" = 'comment on column only in key schema',
			VALUE DOC ON COLUMN TYPE "database"."schema"."table"."c2" = 'comment on column only in value schema'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.Size("xsmall")
		b.From(IdentifierSchemaStruct{Name: "table", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("test_avro_topic")
		b.Key([]string{"key_1", "key_2"})
		b.Format(SinkFormatSpecStruct{Avro: &SinkAvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "public"}}})
		b.Envelope(KafkaSinkEnvelopeStruct{Upsert: true})
		b.Snapshot(false)
		b.AvroDoc("top-level comment")
		b.AvroColumnDoc(
			[]AvroColumnStruct{
				{
					Key:    "comment on column only in key schema",
					Value:  "comment on column only in value schema",
					Column: "c1",
				},
				{
					Key:    "comment on column only in key schema",
					Value:  "comment on column only in value schema",
					Column: "c2",
				},
			},
		)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
