package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestNetworkPolicyDatasource(t *testing.T) {
	r := require.New(t)

	// Empty input map since we're not filtering
	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, NetworkPolicy().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// No predicates since we're listing all policies
		testhelpers.MockNetworkPolicyScan(mock, "")

		if err := networkPolicyRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Verify the data source output
		policies := d.Get("network_policies").([]interface{})
		r.Equal(1, len(policies))

		policy := policies[0].(map[string]interface{})
		r.Equal("u1", policy["id"])
		r.Equal("office_policy", policy["name"])
		r.Equal("Network policy for office locations", policy["comment"])

		rules := policy["rules"].([]interface{})
		r.Equal(2, len(rules))

		rule1 := rules[0].(map[string]interface{})
		r.Equal("minnesota", rule1["name"])
		r.Equal("allow", rule1["action"])
		r.Equal("ingress", rule1["direction"])
		r.Equal("2.3.4.5/32", rule1["address"])

		rule2 := rules[1].(map[string]interface{})
		r.Equal("new_york", rule2["name"])
		r.Equal("allow", rule2["action"])
		r.Equal("ingress", rule2["direction"])
		r.Equal("1.2.3.4/28", rule2["address"])
	})
}
