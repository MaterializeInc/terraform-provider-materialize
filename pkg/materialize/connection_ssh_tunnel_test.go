package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestConnectionSshTunnelCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE CONNECTION "database"."schema"."ssh_conn" TO SSH TUNNEL \(HOST 'localhost', USER 'user', PORT 123\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "ssh_conn", SchemaName: "schema", DatabaseName: "database"}
		b := NewConnectionSshTunnelBuilder(db, o)
		b.SSHHost("localhost")
		b.SSHPort(123)
		b.SSHUser("user")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
