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
			WITH \( ENDPOINT = 'localhost', 
				REGION = 'us-east-1', 
				ACCESS KEY ID = 'foo',
				SECRET ACCESS KEY = SECRET "database"."schema"."password",
				SESSION TOKEN = 'biz',
				ASSUME ROLE ARN = 'arn:aws:iam::123456789012:user/JohnDoe', 
				ASSUME ROLE SESSION NAME = 's3-access-example', 
				VALIDATE = false\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "aws_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionAwsBuilder(db, o)
		b.Endpoint("localhost")
		b.Region("us-east-1")
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
