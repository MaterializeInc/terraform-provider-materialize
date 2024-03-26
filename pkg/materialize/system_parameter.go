package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SystemParameterBuilder struct {
	ddl        Builder
	paramName  string
	paramValue string
}

func NewSystemParameterBuilder(conn *sqlx.DB, paramName, paramValue string) *SystemParameterBuilder {
	return &SystemParameterBuilder{
		ddl:        Builder{conn, System},
		paramName:  paramName,
		paramValue: paramValue,
	}
}

func (b *SystemParameterBuilder) Set() error {
	q := fmt.Sprintf(`ALTER SYSTEM SET "%s" TO '%s';`, b.paramName, b.paramValue)
	return b.ddl.exec(q)
}

func (b *SystemParameterBuilder) Reset() error {
	q := fmt.Sprintf(`ALTER SYSTEM RESET "%s";`, b.paramName)
	return b.ddl.exec(q)
}

func ShowSystemParameter(conn *sqlx.DB, paramName string) (string, error) {
	var paramValue string
	query := fmt.Sprintf(`SHOW %s;`, QuoteIdentifier(paramName))
	err := conn.QueryRow(query).Scan(&paramValue)
	if err != nil {
		return "", fmt.Errorf("error reading system parameter %s: %v", paramName, err)
	}
	return paramValue, nil
}
