package materialize

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
)

type EntityType string
type RawSQL string

const (
	ClusterReplica   EntityType = "CLUSTER REPLICA"
	Cluster          EntityType = "CLUSTER"
	BaseConnection   EntityType = "CONNECTION"
	Database         EntityType = "DATABASE"
	Index            EntityType = "INDEX"
	MaterializedView EntityType = "MATERIALIZED VIEW"
	Privilege        EntityType = "PRIVILEGE"
	Ownership        EntityType = "OWNERSHIP"
	Role             EntityType = "ROLE"
	Schema           EntityType = "SCHEMA"
	BaseSink         EntityType = "SINK"
	BaseSource       EntityType = "SOURCE"
	Secret           EntityType = "SECRET"
	Table            EntityType = "TABLE"
	BaseType         EntityType = "TYPE"
	View             EntityType = "VIEW"
	System           EntityType = "SYSTEM"
)

type Builder struct {
	conn   *sqlx.DB
	entity EntityType
}

func (b *Builder) exec(statement string) error {
	if statement[len(statement)-1:] != ";" {
		statement += ";"
	}

	_, err := b.conn.Exec(statement)
	if err != nil {
		log.Printf("[DEBUG] error executing: %s", statement)
		var pgErr pgx.PgError
		pgErr, ok := err.(pgx.PgError)
		if ok {
			msg := fmt.Sprintf("%s: %s", pgErr.Severity, pgErr.Message)
			if pgErr.Detail != "" {
				msg += fmt.Sprintf(" DETAIL: %s", pgErr.Detail)
			}
			if pgErr.Hint != "" {
				msg += fmt.Sprintf(" HINT: %s", pgErr.Hint)
			}
			msg += fmt.Sprintf(" (SQLSTATE %s)", pgErr.SQLState())
			return errors.New(msg)
		}
		return err
	}

	return nil
}

func (b *Builder) drop(name string) error {
	q := fmt.Sprintf(`DROP %s %s;`, b.entity, name)
	return b.exec(q)
}

func (b *Builder) rename(oldName, newName string) error {
	q := fmt.Sprintf(`ALTER %s %s RENAME TO %s;`, b.entity, oldName, newName)
	return b.exec(q)
}

func (b *Builder) resize(name, size string) error {
	q := fmt.Sprintf(`ALTER %s %s SET (SIZE = '%s');`, b.entity, name, size)
	return b.exec(q)
}

func (b *Builder) alter(name string, setOptions map[string]interface{}, resetOptions []string, isSecret, validate bool) error {
	var clauses []string
	for option, val := range setOptions {
		var setValue string
		switch v := val.(type) {
		case ValueSecretStruct:
			if v.Text != "" {
				setValue = QuoteString(v.Text)
			} else if v.Secret.Name != "" {
				setValue = "SECRET " + v.Secret.QualifiedName()
			}
		case IdentifierSchemaStruct:
			prefix := ""
			if isSecret || option == "SASL PASSWORD" || option == "SSL KEY" {
				prefix = "SECRET "
			}
			setValue = prefix + v.QualifiedName()
		case string:
			setValue = QuoteString(v)
		case int:
			setValue = fmt.Sprintf("%d", v)
		case []string:
			setValue = fmt.Sprintf("[%s]", "'"+strings.Join(v, "', '")+"'")
		case RawSQL:
			setValue = string(v)
		default:
			return fmt.Errorf("unsupported value type for option %s: %T", option, val)
		}

		if setValue == "" {
			return fmt.Errorf("no valid value provided for option %s", option)
		}

		clauses = append(clauses, fmt.Sprintf("SET (%s = %s)", option, setValue))
	}

	for _, option := range resetOptions {
		clauses = append(clauses, fmt.Sprintf("RESET (%s)", option))
	}

	validateClause := ""
	if !validate {
		validateClause = " WITH (validate false)"
	}

	clauseString := strings.Join(clauses, ", ")
	query := fmt.Sprintf(`ALTER %s %s %s%s;`, b.entity, name, clauseString, validateClause)

	return b.exec(query)
}

func (b *Builder) alterDrop(name string, options []string, validate bool) error {
	validateClause := ""
	if !validate {
		validateClause = " WITH (validate false)"
	}

	// Construct each DROP clause separately
	dropClauses := []string{}
	for _, option := range options {
		dropClause := fmt.Sprintf("DROP( %s )", option)
		dropClauses = append(dropClauses, dropClause)
	}

	// Join the DROP clauses with commas
	optionsClause := strings.Join(dropClauses, ", ")

	// Construct the final query
	query := fmt.Sprintf(`ALTER %s %s %s%s;`, b.entity, name, optionsClause, validateClause)

	return b.exec(query)
}
