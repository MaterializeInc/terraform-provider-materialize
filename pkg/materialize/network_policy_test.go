package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestNetworkPolicyCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE NETWORK POLICY "office_policy" \( RULES \( ` +
				`"new_york" \(action='allow', direction='ingress', address='1\.2\.3\.4/28'\), ` +
				`"minnesota" \(action='allow', direction='ingress', address='2\.3\.4\.5/32'\) \)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "office_policy"}
		b := NewNetworkPolicyBuilder(db, o)

		rules := []NetworkPolicyRule{
			{
				Name:      "new_york",
				Action:    "allow",
				Direction: "ingress",
				Address:   "1.2.3.4/28",
			},
			{
				Name:      "minnesota",
				Action:    "allow",
				Direction: "ingress",
				Address:   "2.3.4.5/32",
			},
		}
		b.Rules(rules)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNetworkPolicyCreateNoRules(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE NETWORK POLICY "empty_policy";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "empty_policy"}
		b := NewNetworkPolicyBuilder(db, o)

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNetworkPolicyAlter(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER NETWORK POLICY "office_policy" SET \( RULES \( ` +
				`"new_york" \(action='allow', direction='ingress', address='1\.2\.3\.4/28'\), ` +
				`"boston" \(action='allow', direction='ingress', address='5\.6\.7\.8/24'\) \)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "office_policy"}
		b := NewNetworkPolicyBuilder(db, o)

		rules := []NetworkPolicyRule{
			{
				Name:      "new_york",
				Action:    "allow",
				Direction: "ingress",
				Address:   "1.2.3.4/28",
			},
			{
				Name:      "boston",
				Action:    "allow",
				Direction: "ingress",
				Address:   "5.6.7.8/24",
			},
		}
		b.Rules(rules)

		if err := b.Alter(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNetworkPolicyDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP NETWORK POLICY "office_policy";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "office_policy"}
		if err := NewNetworkPolicyBuilder(db, o).Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestNetworkPolicyScan(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{
			"id", "policy_name", "comment", "owner_name", "privileges", "rules",
		}).AddRow(
			"u1",
			"office_policy",
			"Network policy for office locations",
			"mz_system",
			StringArray{"s1=U/s1"},
			`[{"rule_name":"new_york","rule_action":"allow","rule_direction":"ingress","rule_address":"1.2.3.4/28"},
			  {"rule_name":"minnesota","rule_action":"allow","rule_direction":"ingress","rule_address":"2.3.4.5/32"}]`,
		)

		mock.ExpectQuery(`WITH policy AS`).WillReturnRows(rows)

		policy, err := ScanNetworkPolicy(db, "u1")
		if err != nil {
			t.Fatal(err)
		}

		// Verify policy details
		if policy.PolicyId.String != "u1" {
			t.Errorf("expected policy id u1, got %s", policy.PolicyId.String)
		}
		if policy.PolicyName.String != "office_policy" {
			t.Errorf("expected policy name office_policy, got %s", policy.PolicyName.String)
		}
		if policy.Comment.String != "Network policy for office locations" {
			t.Errorf("expected comment 'Network policy for office locations', got %s", policy.Comment.String)
		}
		if policy.OwnerName.String != "mz_system" {
			t.Errorf("expected owner mz_system, got %s", policy.OwnerName.String)
		}
		if len(policy.Privileges) != 1 || policy.Privileges[0] != "s1=U/s1" {
			t.Errorf("expected privileges [s1=U/s1], got %v", policy.Privileges)
		}

		// Verify rules
		if len(policy.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(policy.Rules))
		}

		expectedRules := []NetworkPolicyRule{
			{
				Name:      "new_york",
				Action:    "allow",
				Direction: "ingress",
				Address:   "1.2.3.4/28",
			},
			{
				Name:      "minnesota",
				Action:    "allow",
				Direction: "ingress",
				Address:   "2.3.4.5/32",
			},
		}

		for i, rule := range policy.Rules {
			if rule != expectedRules[i] {
				t.Errorf("rule %d mismatch: expected %v, got %v", i, expectedRules[i], rule)
			}
		}
	})
}
