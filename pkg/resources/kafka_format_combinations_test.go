package resources

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

// Materialize takes either FORMAT or KEY FORMAT ... VALUE FORMAT, and the key
// and value formats only together. Shared by the Kafka source and Kafka source
// table resources so their plan-time rules cannot drift apart.
func testKafkaFormatCombinations(t *testing.T, r *schema.Resource, base map[string]interface{}) {
	text := []interface{}{map[string]interface{}{"text": true}}
	validate := func(formats map[string]interface{}) error {
		cfg := map[string]interface{}{}
		for k, v := range base {
			cfg[k] = v
		}
		for k, v := range formats {
			cfg[k] = v
		}
		// the format attributes are deprecated, so only errors count here
		var errs []string
		for _, d := range r.Validate(terraform.NewResourceConfigRaw(cfg)) {
			if d.Severity == diag.Error {
				errs = append(errs, d.Summary+": "+d.Detail)
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	cases := []struct {
		name    string
		formats map[string]interface{}
		wantErr string
	}{
		{"no format", nil, ""},
		{"format only", map[string]interface{}{"format": text}, ""},
		{"key and value format", map[string]interface{}{"key_format": text, "value_format": text}, ""},
		{"format with key_format", map[string]interface{}{"format": text, "key_format": text}, "conflicts with"},
		{"format with value_format", map[string]interface{}{"format": text, "value_format": text}, "conflicts with"},
		{"format with both", map[string]interface{}{"format": text, "key_format": text, "value_format": text}, "conflicts with"},
		{"key_format only", map[string]interface{}{"key_format": text}, "all of `key_format,value_format` must be specified"},
		{"value_format only", map[string]interface{}{"value_format": text}, "all of `key_format,value_format` must be specified"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validate(c.formats)
			if c.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), c.wantErr)
		})
	}
}
