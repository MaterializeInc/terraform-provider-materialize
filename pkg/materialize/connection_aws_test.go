package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionAwsCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."aws_conn" TO AWS
			    \( ENDPOINT = 'localhost',
				REGION = 'us-east-1',
				ACCESS KEY ID = 'foo',
				SECRET ACCESS KEY = SECRET "database"."schema"."password",
				SESSION TOKEN = 'biz',
				ASSUME ROLE ARN = 'arn:aws:iam::123456789012:user/JohnDoe',
				ASSUME ROLE SESSION NAME = 's3-access-example'\)
				WITH \(VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionAwsBuilder(db, o)
		b.Endpoint("localhost")
		b.AwsRegion("us-east-1")
		b.AccessKeyId(ValueSecretStruct{Text: "foo"})
		b.SecretAccessKey(IdentifierSchemaStruct{Name: "password", DatabaseName: "database", SchemaName: "schema"})
		b.SessionToken(ValueSecretStruct{Text: "biz"})
		b.AssumeRoleArn("arn:aws:iam::123456789012:user/JohnDoe")
		b.AssumeRoleSessionName("s3-access-example")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestScanConnectionAws(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedExternalId := "mz_12345678-1234-1234-1234-123456789012_u123"

		// Mock the scan query response
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		// Test ScanConnectionAws
		params, err := ScanConnectionAws(db, "u1")
		if err != nil {
			t.Fatal(err)
		}

		if params.ExternalId.String != expectedExternalId {
			t.Fatalf("Expected external_id to be %s, got %s", expectedExternalId, params.ExternalId.String)
		}

		if params.AssumeRoleArn.String != "arn:aws:iam::123456789012:user/JohnDoe" {
			t.Fatalf("Expected assume_role_arn to be %s, got %s", "arn:aws:iam::123456789012:user/JohnDoe", params.AssumeRoleArn.String)
		}
	})
}
