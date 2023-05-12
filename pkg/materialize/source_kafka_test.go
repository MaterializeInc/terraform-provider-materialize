package materialize

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestResourceSourceKafkaCreate(t *testing.T) {
// 	r := require.New(t)
// 	b := NewSourceKafkaBuilder("source", "schema", "database")
// 	b.Size("xsmall")
// 	b.KafkaConnection(IdentifierSchemaStruct{Name: "kafka_connection", DatabaseName: "database", SchemaName: "schema"})
// 	b.Topic("events")
// 	b.Format(FormatSpecStruct{Avro: &AvroFormatSpec{SchemaRegistryConnection: IdentifierSchemaStruct{Name: "csr_connection", DatabaseName: "database", SchemaName: "schema"}}})
// 	b.IncludeKey()
// 	b.IncludeHeaders()
// 	b.IncludePartition()
// 	b.IncludeOffset()
// 	b.IncludeTimestamp()
// 	b.Envelope(KafkaSourceEnvelopeStruct{Upsert: true})
// 	r.Equal(`CREATE SOURCE "database"."schema"."source" FROM KAFKA CONNECTION "database"."schema"."kafka_connection" (TOPIC 'events') FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection" INCLUDE KEY, HEADERS, PARTITION, OFFSET, TIMESTAMP ENVELOPE UPSERT WITH (SIZE = 'xsmall');`, b.Create())
// }
