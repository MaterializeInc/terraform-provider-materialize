package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inNetworkPolicy = map[string]interface{}{
	"name": "office_policy",
	"rule": []interface{}{
		map[string]interface{}{
			"name":      "minnesota",
			"action":    "allow",
			"direction": "ingress",
			"address":   "2.3.4.5/32",
		},
		map[string]interface{}{
			"name":      "new_york",
			"action":    "allow",
			"direction": "ingress",
			"address":   "1.2.3.4/28",
		},
	},
	"comment": "Network policy for office locations",
}

func TestResourceNetworkPolicyCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, NetworkPolicy().Schema, inNetworkPolicy)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE NETWORK POLICY "office_policy" \( RULES \( "minnesota" \(action='allow', direction='ingress', address='2\.3\.4\.5/32'\), "new_york" \(action='allow', direction='ingress', address='1\.2\.3\.4/28'\) \)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(
			`COMMENT ON NETWORK POLICY "office_policy" IS 'Network policy for office locations';`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE policy_name = 'office_policy'`
		testhelpers.MockNetworkPolicyScan(mock, ip)

		// Query Params
		pp := `WHERE policy.id = 'u1'`
		testhelpers.MockNetworkPolicyScan(mock, pp)

		if err := networkPolicyCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceNetworkPolicyReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, NetworkPolicy().Schema, inNetworkPolicy)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		pp := `WHERE policy.id = 'u1'`
		testhelpers.MockNetworkPolicyScan(mock, pp)

		if err := networkPolicyRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceNetworkPolicyUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, NetworkPolicy().Schema, map[string]interface{}{
		"name": "office_policy",
		"rule": []interface{}{
			map[string]interface{}{
				"name":      "boston",
				"action":    "allow",
				"direction": "ingress",
				"address":   "5.6.7.8/24",
			},
			map[string]interface{}{
				"name":      "new_york",
				"action":    "allow",
				"direction": "ingress",
				"address":   "1.2.3.4/28",
			},
		},
	})
	r.NotNil(d)
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Alter
		mock.ExpectExec(
			`ALTER NETWORK POLICY "office_policy" SET \( RULES \( "boston" \(action='allow', direction='ingress', address='5\.6\.7\.8/24'\), "new_york" \(action='allow', direction='ingress', address='1\.2\.3\.4/28'\) \)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query after update
		pp := `WHERE policy.id = 'u1'`
		testhelpers.MockNetworkPolicyScan(mock, pp)

		if err := networkPolicyUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceNetworkPolicyDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, NetworkPolicy().Schema, inNetworkPolicy)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP NETWORK POLICY "office_policy";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := networkPolicyDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
