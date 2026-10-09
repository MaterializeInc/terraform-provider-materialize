package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Materialize rejects an upsert or Debezium envelope without a key format,
// unless the value format is Avro or Protobuf, where the key schema comes from
// the schema registry. Catch it in the plan instead of failing at apply.
func validateKafkaEnvelopeKeyFormat(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	for _, k := range []string{"envelope", "key_format", "format"} {
		if !d.NewValueKnown(k) {
			return nil
		}
	}

	envelopes, _ := d.Get("envelope").([]interface{})
	if len(envelopes) == 0 || envelopes[0] == nil {
		return nil
	}
	envelope, _ := envelopes[0].(map[string]interface{})
	var name string
	if upsert, _ := envelope["upsert"].(bool); upsert {
		name = "upsert"
	} else if debezium, _ := envelope["debezium"].(bool); debezium {
		name = "debezium"
	} else {
		return nil
	}

	if keyFormat, _ := d.Get("key_format").([]interface{}); len(keyFormat) > 0 {
		return nil
	}
	if formats, _ := d.Get("format").([]interface{}); len(formats) > 0 && formats[0] != nil {
		format, _ := formats[0].(map[string]interface{})
		for _, registry := range []string{"avro", "protobuf"} {
			if f, _ := format[registry].([]interface{}); len(f) > 0 {
				return nil
			}
		}
	}

	return fmt.Errorf("envelope %s requires key_format and value_format, unless format uses avro or protobuf, which take the key schema from the schema registry", name)
}
