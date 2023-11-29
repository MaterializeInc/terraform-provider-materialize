package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://github.com/MaterializeInc/materialize/blob/main/test/testdrive/kafka-sinks.td
// https://github.com/MaterializeInc/materialize/blob/main/test/testdrive/kafka-json-sinks.td
// https://materialize.com/docs/sql/create-sink/kafka/

func TestSinkKafkaCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn"
			\(TOPIC 'testdrive-snk1-seed'\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."csr_conn"
			ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("testdrive-snk1-seed")
		b.Format(
			SinkFormatSpecStruct{
				Avro: &SinkAvroFormatSpec{
					SchemaRegistryConnection: IdentifierSchemaStruct{
						Name:         "csr_conn",
						DatabaseName: "materialize",
						SchemaName:   "public",
					},
				},
			},
		)
		b.Envelope(KafkaSinkEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaSnapshotCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'testdrive-snk1-seed'\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."csr_conn"
			ENVELOPE DEBEZIUM
			WITH \(SNAPSHOT = true\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("testdrive-snk1-seed")
		b.Format(
			SinkFormatSpecStruct{
				Avro: &SinkAvroFormatSpec{
					SchemaRegistryConnection: IdentifierSchemaStruct{
						Name:         "csr_conn",
						DatabaseName: "materialize",
						SchemaName:   "public",
					},
				},
			},
		)
		b.Snapshot(true)
		b.Envelope(KafkaSinkEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaSizeSnapshotCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'testdrive-snk1-seed'\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "materialize"."public"."csr_conn"
			ENVELOPE DEBEZIUM
			WITH \(SIZE = '2xsmall', SNAPSHOT = true\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("testdrive-snk1-seed")
		b.Format(
			SinkFormatSpecStruct{
				Avro: &SinkAvroFormatSpec{
					SchemaRegistryConnection: IdentifierSchemaStruct{
						Name:         "csr_conn",
						DatabaseName: "materialize",
						SchemaName:   "public",
					},
				},
			},
		)
		b.Size("2xsmall")
		b.Snapshot(true)
		b.Envelope(KafkaSinkEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaJsonCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'testdrive-snk1-seed'\)
			FORMAT JSON
			ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("testdrive-snk1-seed")
		b.Format(SinkFormatSpecStruct{Json: true})
		b.Envelope(KafkaSinkEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaKeyCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'testdrive-snk1-seed'\)
			KEY \(b\)
			FORMAT JSON
			ENVELOPE DEBEZIUM;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.Topic("testdrive-snk1-seed")
		b.Format(SinkFormatSpecStruct{Json: true})
		b.Key([]string{"b"})
		b.Envelope(KafkaSinkEnvelopeStruct{Debezium: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaKeyNotEnforcedCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			IN CLUSTER "my_io_cluster"
			FROM "database"."schema"."src"
			INTO KAFKA CONNECTION "database"."schema"."kafka_conn" \(TOPIC 'testdrive-snk1-seed'\)
			KEY \(k\) NOT ENFORCED
			FORMAT JSON
			ENVELOPE UPSERT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "sink", SchemaName: "schema", DatabaseName: "database"}
		b := NewSinkKafkaBuilder(db, o)
		b.From(IdentifierSchemaStruct{Name: "src", SchemaName: "schema", DatabaseName: "database"})
		b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_conn", SchemaName: "schema", DatabaseName: "database"})
		b.ClusterName("my_io_cluster")
		b.Topic("testdrive-snk1-seed")
		b.Format(SinkFormatSpecStruct{Json: true})
		b.Key([]string{"k"})
		b.KeyNotEnforced(true)
		b.Envelope(KafkaSinkEnvelopeStruct{Upsert: true})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSinkKafkaAvroDocsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."sink"
			FROM "database"."schema"."table"
			INTO KAFKA CONNECTION "database"."schema"."kafka_connection"
			\(TOPIC 'testdrive-snk1-seed'\)
			FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."public"."csr_connection"
			ENVELOPE UPSERT WITH \(SIZE = 'xsmall'\)
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
		b.KafkaConnection(IdentifierSchemaStruct{
			Name:         "kafka_connection",
			SchemaName:   "schema",
			DatabaseName: "database",
		})
		b.Topic("testdrive-snk1-seed")
		b.Format(SinkFormatSpecStruct{
			Avro: &SinkAvroFormatSpec{
				SchemaRegistryConnection: IdentifierSchemaStruct{
					Name:         "csr_connection",
					DatabaseName: "database",
					SchemaName:   "public",
				},
			},
		})
		b.Envelope(KafkaSinkEnvelopeStruct{Upsert: true})
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
