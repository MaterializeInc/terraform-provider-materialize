package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionAwsPrivatelinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."privatelink_conn" TO AWS PRIVATELINK \(SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES \('use1-az1', 'use1-az2'\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := ObjectSchemaStruct{Name: "privatelink_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionAwsPrivatelinkBuilder(db, o)
		b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
		b.PrivateLinkAvailabilityZones([]string{"use1-az1", "use1-az2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
