package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestRoleParameterSet(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		roleName := "test_role"
		variableName := "transaction_isolation"
		variableValue := "strict serializable"

		mock.ExpectExec(`ALTER ROLE "` + roleName + `" SET "` + variableName + `" TO '` + variableValue + `';`).WillReturnResult(sqlmock.NewResult(0, 1))

		rpBuilder := NewRoleParameterBuilder(db, roleName, variableName, variableValue)
		if err := rpBuilder.Set(); err != nil {
			t.Fatalf("unexpected error during Set: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestRoleParameterReset(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		roleName := "test_role"
		variableName := "transaction_isolation"

		mock.ExpectExec(`ALTER ROLE "` + roleName + `" RESET "` + variableName + `";`).WillReturnResult(sqlmock.NewResult(0, 1))

		rpBuilder := NewRoleParameterBuilder(db, roleName, variableName, "")
		if err := rpBuilder.Reset(); err != nil {
			t.Fatalf("unexpected error during Reset: %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
