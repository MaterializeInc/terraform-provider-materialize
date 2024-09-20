package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestResourceSourceTableKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."source"
            FROM SOURCE "database"."schema"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT JSON
            INCLUDE KEY AS message_key, HEADERS AS message_headers, PARTITION AS message_partition
            ENVELOPE UPSERT
            EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
		b := NewSourceTableKafkaBuilder(db, o)
		b.Source(IdentifierSchemaStruct{Name: "kafka_source", DatabaseName: "database", SchemaName: "schema"})
		b.UpstreamName("topic")
		b.Format(SourceFormatSpecStruct{Json: true})
		b.IncludeKey()
		b.IncludeKeyAlias("message_key")
		b.IncludeHeaders()
		b.IncludeHeadersAlias("message_headers")
		b.IncludePartition()
		b.IncludePartitionAlias("message_partition")
		b.Envelope(KafkaSourceEnvelopeStruct{Upsert: true})
		b.ExposeProgress(IdentifierSchemaStruct{Name: "progress", DatabaseName: "database", SchemaName: "schema"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithAvroFormat(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."source"
            FROM SOURCE "database"."schema"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."schema_registry"
            KEY STRATEGY EXTRACT
            VALUE STRATEGY EXTRACT
            INCLUDE TIMESTAMP
            ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
		b := NewSourceTableKafkaBuilder(db, o)
		b.Source(IdentifierSchemaStruct{Name: "kafka_source", DatabaseName: "database", SchemaName: "schema"})
		b.UpstreamName("topic")
		b.Format(SourceFormatSpecStruct{
			Avro: &AvroFormatSpec{
				SchemaRegistryConnection: IdentifierSchemaStruct{Name: "schema_registry", DatabaseName: "database", SchemaName: "schema"},
				KeyStrategy:              "EXTRACT",
				ValueStrategy:            "EXTRACT",
			},
		})
		b.IncludeTimestamp()
		b.Envelope(KafkaSourceEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableKafkaCreateWithUpsertOptions(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."source"
            FROM SOURCE "database"."schema"."kafka_source"
            \(REFERENCE "topic"\)
            FORMAT JSON
            INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP
            ENVELOPE UPSERT \(VALUE DECODING ERRORS = \(INLINE AS my_error_col\)\)
            EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "source", SchemaName: "schema", DatabaseName: "database"}
		b := NewSourceTableKafkaBuilder(db, o)
		b.Source(IdentifierSchemaStruct{Name: "kafka_source", DatabaseName: "database", SchemaName: "schema"})
		b.UpstreamName("topic")
		b.Format(SourceFormatSpecStruct{Json: true})
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
