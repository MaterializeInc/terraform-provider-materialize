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
			`CREATE SOURCE "database"."schema"."source"
            FROM KAFKA CONNECTION "database"."schema"."kafka_connection"
            \(TOPIC 'events'\) FORMAT AVRO
            USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection"
            INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT
            EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
		b := NewSourceKafkaBuilder(db, o)
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
		b.Topic("events")
		b.Format(SourceFormatSpecStruct{Avro: &AvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "schema"}}})
		b.IncludeKey()
		b.IncludeHeaders()
		b.IncludePartition()
		b.IncludeOffset()
		b.IncludeTimestamp()
		b.Envelope(KafkaSourceEnvelopeStruct{Upsert: true})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceKafkaCreateWithUpsertOptions(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source"
            FROM KAFKA CONNECTION "database"."schema"."kafka_connection"
            \(TOPIC 'events'\) FORMAT AVRO
            USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection"
            INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT
            \(VALUE DECODING ERRORS = \(INLINE AS my_error_col\)\)
            EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
		b := NewSourceKafkaBuilder(db, o)
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
		b.Topic("events")
		b.Format(SourceFormatSpecStruct{Avro: &AvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "schema"}}})
		b.IncludeKey()
		b.IncludeHeaders()
		b.IncludePartition()
		b.IncludeOffset()
		b.IncludeTimestamp()
		b.Envelope(KafkaSourceEnvelopeStruct{
			Upsert: true,
			UpsertOptions: &UpsertOptionsStruct{
				ValueDecodingErrors: struct {
					Inline struct {
						Enabled bool
						Alias   string
					}
				}{
					Inline: struct {
						Enabled bool
						Alias   string
					}{
						Enabled: true,
						Alias:   "my_error_col",
					},
				},
			},
		})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
