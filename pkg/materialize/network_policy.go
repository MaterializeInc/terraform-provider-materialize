package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type NetworkPolicyRule struct {
	Name      string
	Action    string
	Direction string
	Address   string
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
			ruleStrings[i] = fmt.Sprintf(`%s (action='%s', direction='%s', address='%s')`,
				QuoteIdentifier(rule.Name),
				rule.Action,
				rule.Direction,
				rule.Address)
		}
		q.WriteString(strings.Join(ruleStrings, ", "))
		q.WriteString(` ))`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *NetworkPolicyBuilder) Alter() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`ALTER NETWORK POLICY %s`, QuoteIdentifier(b.name)))

	if len(b.rules) > 0 {
		q.WriteString(` ( RULES ( `)
		ruleStrings := make([]string, len(b.rules))
		for i, rule := range b.rules {
			ruleStrings[i] = fmt.Sprintf(`%s (action='%s', direction='%s', address='%s')`,
				QuoteIdentifier(rule.Name),
				rule.Action,
				rule.Direction,
				rule.Address)
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
}

var networkPolicyQuery = NewBaseQuery(`
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
	) comments
		ON mz_network_policies.id = comments.id`)

func NetworkPolicyId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_network_policies.name": obj.Name,
	}
	q := networkPolicyQuery.QueryPredicate(p)

	var c NetworkPolicyParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.PolicyId.String, nil
}

func ScanNetworkPolicy(conn *sqlx.DB, id string) (NetworkPolicyParams, error) {
	p := map[string]string{
		"mz_network_policies.id": id,
	}
	q := networkPolicyQuery.QueryPredicate(p)

	var c NetworkPolicyParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
