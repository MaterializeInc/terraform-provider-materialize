package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

// Mirrors what Materialize v26.43 accepts: an upsert or Debezium envelope needs
// a key format unless the value format is Avro or Protobuf, which take the key
// schema from the schema registry.
func TestKafkaEnvelopeRequiresKeyFormat(t *testing.T) {
	text := []interface{}{map[string]interface{}{"text": true}}
	json := []interface{}{map[string]interface{}{"json": true}}
	csr := []interface{}{map[string]interface{}{"name": "csr_conn"}}
	avro := []interface{}{map[string]interface{}{"avro": []interface{}{map[string]interface{}{"schema_registry_connection": csr}}}}
	protobuf := []interface{}{map[string]interface{}{"protobuf": []interface{}{map[string]interface{}{"schema_registry_connection": csr, "message": "Msg"}}}}
	envelope := func(kind string) []interface{} { return []interface{}{map[string]interface{}{kind: true}} }

	cases := []struct {
		name    string
		attrs   map[string]interface{}
		wantErr string
	}{
		{"no envelope", map[string]interface{}{"format": json}, ""},
		{"envelope none", map[string]interface{}{"format": json, "envelope": envelope("none")}, ""},
		{"upsert with key and value format", map[string]interface{}{"key_format": text, "value_format": json, "envelope": envelope("upsert")}, ""},
		{"upsert with avro", map[string]interface{}{"format": avro, "envelope": envelope("upsert")}, ""},
		{"upsert with protobuf", map[string]interface{}{"format": protobuf, "envelope": envelope("upsert")}, ""},
		{"debezium with avro", map[string]interface{}{"format": avro, "envelope": envelope("debezium")}, ""},
		{"upsert with json", map[string]interface{}{"format": json, "envelope": envelope("upsert")}, "envelope upsert requires key_format"},
		{"upsert with text", map[string]interface{}{"format": text, "envelope": envelope("upsert")}, "envelope upsert requires key_format"},
		{"upsert without any format", map[string]interface{}{"envelope": envelope("upsert")}, "envelope upsert requires key_format"},
		{"debezium with json", map[string]interface{}{"format": json, "envelope": envelope("debezium")}, "envelope debezium requires key_format"},
	}

	resources := map[string]struct {
		resource *schema.Resource
		base     map[string]interface{}
	}{
		"materialize_source_kafka": {SourceKafka(), map[string]interface{}{
			"name": "source", "topic": "topic",
			"kafka_connection": []interface{}{map[string]interface{}{"name": "kafka_conn"}},
		}},
		"materialize_source_table_kafka": {SourceTableKafka(), map[string]interface{}{
			"name": "table", "topic": "topic",
			"source": []interface{}{map[string]interface{}{"name": "kafka_source"}},
		}},
	}

	for rname, r := range resources {
		for _, c := range cases {
			t.Run(rname+"/"+c.name, func(t *testing.T) {
				cfg := map[string]interface{}{}
				for k, v := range r.base {
					cfg[k] = v
				}
				for k, v := range c.attrs {
					cfg[k] = v
				}
				_, err := r.resource.Diff(context.TODO(), nil, terraform.NewResourceConfigRaw(cfg), nil)
				if c.wantErr == "" {
					require.NoError(t, err)
					return
				}
				require.Error(t, err)
				require.Contains(t, err.Error(), c.wantErr)
			})
		}
	}
}
