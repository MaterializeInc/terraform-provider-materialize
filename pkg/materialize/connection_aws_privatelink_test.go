package materialize

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionAwsPrivateLinkCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."privatelink_conn" TO AWS PRIVATELINK \(SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',AVAILABILITY ZONES \('use1-az1', 'use1-az2'\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewConnectionAwsPrivatelinkBuilder(db, "privatelink_conn", "schema", "database")
		b.PrivateLinkServiceName("com.amazonaws.us-east-1.materialize.example")
		b.PrivateLinkAvailabilityZones([]string{"use1-az1", "use1-az2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
