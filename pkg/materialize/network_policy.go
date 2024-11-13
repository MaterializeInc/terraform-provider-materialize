package materialize

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type NetworkPolicyRule struct {
	Name      string `json:"rule_name"`
	Action    string `json:"rule_action"`
	Direction string `json:"rule_direction"`
	Address   string `json:"rule_address"`
}

type NetworkPolicyBuilder struct {
	ddl   Builder
	name  string
	rules []NetworkPolicyRule
}

func NewNetworkPolicyBuilder(conn *sqlx.DB, obj MaterializeObject) *NetworkPolicyBuilder {
	return &NetworkPolicyBuilder{
		ddl:  Builder{conn, NetworkPolicy},
		name: obj.Name,
	}
}

func (b *NetworkPolicyBuilder) Rules(rules []NetworkPolicyRule) *NetworkPolicyBuilder {
	b.rules = rules
	return b
}

func (b *NetworkPolicyBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE NETWORK POLICY %s`, QuoteIdentifier(b.name)))

	if len(b.rules) > 0 {
		q.WriteString(` ( RULES ( `)
		ruleStrings := make([]string, len(b.rules))
		for i, rule := range b.rules {
			ruleStrings[i] = fmt.Sprintf(`%s (action=%s, direction=%s, address=%s)`,
				QuoteIdentifier(rule.Name),
				QuoteString(rule.Action),
				QuoteString(rule.Direction),
				QuoteString(rule.Address))
		}
		q.WriteString(strings.Join(ruleStrings, ", "))
		q.WriteString(` ))`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *NetworkPolicyBuilder) Alter() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`ALTER NETWORK POLICY %s SET`, QuoteIdentifier(b.name)))

	if len(b.rules) > 0 {
		q.WriteString(` ( RULES ( `)
		ruleStrings := make([]string, len(b.rules))
		for i, rule := range b.rules {
			ruleStrings[i] = fmt.Sprintf(`%s (action=%s, direction=%s, address=%s)`,
				QuoteIdentifier(rule.Name),
				QuoteString(rule.Action),
				QuoteString(rule.Direction),
				QuoteString(rule.Address))
		}
		q.WriteString(strings.Join(ruleStrings, ", "))
		q.WriteString(` ))`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *NetworkPolicyBuilder) Drop() error {
	return b.ddl.drop(QuoteIdentifier(b.name))
}

// DML
type NetworkPolicyParams struct {
	PolicyId   sql.NullString `db:"id"`
	PolicyName sql.NullString `db:"policy_name"`
	Comment    sql.NullString `db:"comment"`
	OwnerName  sql.NullString `db:"owner_name"`
	Privileges pq.StringArray `db:"privileges"`
	Rules      []NetworkPolicyRule
}

type NetworkPolicyQueryResult struct {
	PolicyId   sql.NullString `db:"id"`
	PolicyName sql.NullString `db:"policy_name"`
	Comment    sql.NullString `db:"comment"`
	OwnerName  sql.NullString `db:"owner_name"`
	Privileges pq.StringArray `db:"privileges"`
	Rules      []byte         `db:"rules"`
}

var networkPolicyQuery = NewBaseQuery(`
	WITH policy AS (
		SELECT
			mz_network_policies.id,
			mz_network_policies.name AS policy_name,
			comments.comment AS comment,
			mz_roles.name AS owner_name,
			mz_network_policies.privileges
		FROM mz_internal.mz_network_policies
		JOIN mz_roles
			ON mz_network_policies.owner_id = mz_roles.id
		LEFT JOIN (
			SELECT id, comment
			FROM mz_internal.mz_comments
			WHERE object_type = 'network-policy'
		) comments ON mz_network_policies.id = comments.id
	),
	rules AS (
		SELECT
			policy_id,
			jsonb_agg(
				jsonb_build_object(
					'rule_name', name,
					'rule_action', action,
					'rule_direction', direction,
					'rule_address', address
				)
			) as rules
		FROM mz_internal.mz_network_policy_rules
		GROUP BY policy_id
	)
	SELECT
		policy.*,
		COALESCE(rules.rules, '[]'::json) as rules
	FROM policy
	LEFT JOIN rules ON policy.id = rules.policy_id`)

func NetworkPolicyId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"policy_name": obj.Name,
	}
	q := networkPolicyQuery.QueryPredicate(p)

	var result NetworkPolicyQueryResult
	if err := conn.Get(&result, q); err != nil {
		return "", err
	}

	return result.PolicyId.String, nil
}

func ScanNetworkPolicy(conn *sqlx.DB, id string) (NetworkPolicyParams, error) {
	p := map[string]string{
		"policy.id": id,
	}
	q := networkPolicyQuery.QueryPredicate(p)

	var result NetworkPolicyQueryResult
	if err := conn.Get(&result, q); err != nil {
		return NetworkPolicyParams{}, err
	}

	policy := NetworkPolicyParams{
		PolicyId:   result.PolicyId,
		PolicyName: result.PolicyName,
		Comment:    result.Comment,
		OwnerName:  result.OwnerName,
		Privileges: result.Privileges,
	}

	// Parse the JSON rules
	if err := json.Unmarshal(result.Rules, &policy.Rules); err != nil {
		return NetworkPolicyParams{}, err
	}

	return policy, nil
}

func ListNetworkPolicies(conn *sqlx.DB) ([]NetworkPolicyParams, error) {
	var policies []NetworkPolicyParams
	q := networkPolicyQuery.QueryPredicate(map[string]string{})

	var results []NetworkPolicyQueryResult
	if err := conn.Select(&results, q); err != nil {
		return policies, err
	}

	for _, result := range results {
		policy := NetworkPolicyParams{
			PolicyId:   result.PolicyId,
			PolicyName: result.PolicyName,
			Comment:    result.Comment,
			OwnerName:  result.OwnerName,
			Privileges: result.Privileges,
		}

		// Parse the JSON rules
		if err := json.Unmarshal(result.Rules, &policy.Rules); err != nil {
			return policies, err
		}

		policies = append(policies, policy)
	}

	return policies, nil
}
