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

var inAws = map[string]interface{}{
	"name":                     "conn",
	"endpoint":                 "http://localhost:4566",
	"aws_region":               "us-east-1",
	"schema_name":              "schema",
	"database_name":            "database",
	"access_key_id":            []interface{}{map[string]interface{}{"text": "foo"}},
	"secret_access_key":        []interface{}{map[string]interface{}{"name": "conn_secret"}},
	"session_token":            []interface{}{map[string]interface{}{"secret": []interface{}{map[string]interface{}{"name": "conn_session"}}}},
	"assume_role_arn":          "arn:aws:iam::123456789012:role/role",
	"assume_role_session_name": "session",
}

func TestResourceConnectionAwsCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."conn" TO AWS \( ENDPOINT = 'http://localhost:4566', REGION = 'us-east-1', ACCESS KEY ID = 'foo', SECRET ACCESS KEY = SECRET "materialize"."public"."conn_secret", SESSION TOKEN = SECRET "materialize"."public"."conn_session", ASSUME ROLE ARN = 'arn:aws:iam::123456789012:role/role', ASSUME ROLE SESSION NAME = 'session'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_connections.name = 'conn' AND mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockConnectionScan(mock, ip)

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}

// Confirm id is updated with region for 0.4.0
func TestResourceConnectionAwsReadIdMigration(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})

}

func TestResourceConnectionAwsUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionAws().Schema, inAws)

	// Set current state
	d.SetId("u1")
	d.Set("name", "old_conn")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."" RENAME TO "conn";`).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(ENDPOINT = 'http://localhost:4566'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(REGION = 'us-east-1'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(ACCESS KEY ID = 'foo'\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(SECRET ACCESS KEY = SECRET "materialize"."public"."conn_secret"\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(SESSION TOKEN = SECRET "materialize"."public"."conn_session"\);`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`ALTER CONNECTION "database"."schema"."old_conn" SET \(ASSUME ROLE ARN = 'arn:aws:iam::123456789012:role/role'\), SET \(ASSUME ROLE SESSION NAME = 'session'\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		if err := connectionAwsUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
