package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestSystemParameterSet(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		paramName := "max_connections"
		paramValue := "100"

		mock.ExpectExec(`ALTER SYSTEM SET "` + paramName + `" TO '` + paramValue + `';`).WillReturnResult(sqlmock.NewResult(0, 1))

		spBuilder := NewSystemParameterBuilder(db, paramName, paramValue)
		if err := spBuilder.Set(); err != nil {
			t.Fatalf("unexpected error during Set: %v", err)
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestSystemParameterReset(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Define the system parameter name that you want to reset
		paramName := "max_connections"

		// Expect the ALTER SYSTEM RESET query to be executed
		mock.ExpectExec(`ALTER SYSTEM RESET "` + paramName + `";`).WillReturnResult(sqlmock.NewResult(0, 1))

		// Create a new SystemParameterBuilder and attempt to reset the system parameter
		spBuilder := NewSystemParameterBuilder(db, paramName, "")
		if err := spBuilder.Reset(); err != nil {
			t.Fatalf("unexpected error during Reset: %v", err)
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestShowSystemParameter(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Define the system parameter name and the expected value to be retrieved
		paramName := "max_connections"
		expectedValue := "100"

		// Expect the SHOW query to be executed and return the expected value
		mock.ExpectQuery(`SHOW "` + paramName + `";`).WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(expectedValue))

		// Attempt to retrieve the system parameter value
		value, err := ShowSystemParameter(db, paramName)
		if err != nil {
			t.Fatalf("unexpected error during ShowSystemParameter: %v", err)
		}

		// Verify the returned value matches the expected value
		if value != expectedValue {
			t.Errorf("expected value %s, got %s", expectedValue, value)
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
