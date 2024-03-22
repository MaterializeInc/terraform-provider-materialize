package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RoleParameterBuilder struct {
	ddl           Builder
	roleName      string
	variableName  string
	variableValue string
}

func NewRoleParameterBuilder(conn *sqlx.DB, roleName, variableName, variableValue string) *RoleParameterBuilder {
	return &RoleParameterBuilder{
		ddl:           Builder{conn, System},
		roleName:      roleName,
		variableName:  variableName,
		variableValue: variableValue,
	}
}

func (b *RoleParameterBuilder) Set() error {
	q := fmt.Sprintf(`ALTER ROLE "%s" SET "%s" TO '%s';`, b.roleName, b.variableName, b.variableValue)
	return b.ddl.exec(q)
}

func (b *RoleParameterBuilder) Reset() error {
	q := fmt.Sprintf(`ALTER ROLE "%s" RESET "%s";`, b.roleName, b.variableName)
	return b.ddl.exec(q)
}

// TODO: Once possible, implement ShowRoleParameter
func ShowRoleParameter(conn *sqlx.DB, roleName, variableName string) (string, error) {
	var variableValue string
	query := fmt.Sprintf(`SHOW "%s";`, variableName)
	err := conn.QueryRow(query).Scan(&variableValue)
	if err != nil {
		return "", fmt.Errorf("error reading variable %s for role %s: %v", variableName, roleName, err)
	}
	return variableValue, nil
}
