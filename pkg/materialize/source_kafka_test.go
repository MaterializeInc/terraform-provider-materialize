package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestResourceSourceKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION "database"."schema"."kafka_connection" \(TOPIC 'events'\) FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection" INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH \(SIZE = 'xsmall'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceKafkaBuilder(db, "source", "schema", "database")
		b.Size("xsmall")
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
		b.Topic("events")
		b.Format(FormatSpecStruct{Avro: &AvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "schema"}}})
		b.IncludeKey()
		b.IncludeHeaders()
		b.IncludePartition()
		b.IncludeOffset()
		b.IncludeTimestamp()
		b.Envelope(KafkaSourceEnvelopeStruct{Upsert: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
