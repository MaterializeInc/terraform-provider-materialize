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
