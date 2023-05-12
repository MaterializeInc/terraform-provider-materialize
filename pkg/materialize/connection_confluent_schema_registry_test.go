package materialize

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func TestConnectionCreateConfluentSchemaRegistryQuery(t *testing.T) {
// 	r := require.New(t)
// 	b := NewConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
// 	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
// 	b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Text: "user"})
// 	b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
// 	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = 'user', PASSWORD = SECRET "database"."schema"."password");`, b.Create())
// }

// func TestConnectionCreateConfluentSchemaRegistryQueryUsernameSecret(t *testing.T) {
// 	r := require.New(t)
// 	b := NewConnectionConfluentSchemaRegistryBuilder("csr_conn", "schema", "database")
// 	b.ConfluentSchemaRegistryUrl("http://localhost:8081")
// 	b.ConfluentSchemaRegistryUsername(ValueSecretStruct{Secret: IdentifierSchemaStruct{SchemaName: "schema", Name: "user", DatabaseName: "database"}})
// 	b.ConfluentSchemaRegistryPassword(IdentifierSchemaStruct{SchemaName: "schema", Name: "password", DatabaseName: "database"})
// 	r.Equal(`CREATE CONNECTION "database"."schema"."csr_conn" TO CONFLUENT SCHEMA REGISTRY (URL 'http://localhost:8081', USERNAME = SECRET "database"."schema"."user", PASSWORD = SECRET "database"."schema"."password");`, b.Create())
// }
