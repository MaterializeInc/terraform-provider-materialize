package resources

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestResourceDatabaseCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "database",
	}
	d := schema.TestResourceDataRaw(t, Database().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE DATABASE "database";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Drop public schema
		mock.ExpectExec(`DROP SCHEMA IF EXISTS "database"."public";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database'`
		testhelpers.MockDatabaseScan(mock, ip)

		// Query Params
		pp := `WHERE mz_databases.id = 'u1'`
		testhelpers.MockDatabaseScan(mock, pp)

		if err := databaseCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceDatabaseReadIdMigration(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "database",
	}
	d := schema.TestResourceDataRaw(t, Database().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_databases.id = 'u1'`
		testhelpers.MockDatabaseScan(mock, pp)

		if err := databaseRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceDatabaseDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "database",
	}
	d := schema.TestResourceDataRaw(t, Database().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP DATABASE "database";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := databaseDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// The reads match "not found" with errors.Is, so an ErrNoRows that arrives
// wrapped still drops the resource from state instead of failing the plan. The
// bare case is the control: both must behave the same way.
func TestResourceDatabaseReadNotFound(t *testing.T) {
	tests := []struct {
		name    string
		scanErr error
	}{
		{"bare", sql.ErrNoRows},
		{"wrapped", fmt.Errorf("scanning database: %w", sql.ErrNoRows)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)

			d := schema.TestResourceDataRaw(t, Database().Schema, map[string]interface{}{"name": "database"})
			r.NotNil(d)
			d.SetId("aws/us-east-1:u1")

			testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WHERE mz_databases.id = 'u1'`).WillReturnError(tt.scanErr)

				diags := databaseRead(context.TODO(), d, db)
				r.Nil(diags, "a missing database should not be reported as an error")
				r.Empty(d.Id(), "a missing database should be removed from state")
			})
		})
	}
}
