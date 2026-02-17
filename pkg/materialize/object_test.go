package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestObjectName(t *testing.T) {
	r := require.New(t)

	on := MaterializeObject{Name: "name"}
	r.Equal(on.QualifiedName(), `"name"`)

	ond := MaterializeObject{Name: "name", DatabaseName: "database"}
	r.Equal(ond.QualifiedName(), `"database"."name"`)

	onsd := MaterializeObject{Name: "name", SchemaName: "schema", DatabaseName: "database"}
	r.Equal(onsd.QualifiedName(), `"database"."schema"."name"`)

	onc := MaterializeObject{Name: "name", ClusterName: "cluster"}
	r.Equal(onc.QualifiedName(), `"cluster"."name"`)
}

func TestObjectId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		o := MaterializeObject{ObjectType: Database, Name: "materialize"}

		// Query Id
		ip := `WHERE mz_databases.name = 'materialize'`
		testhelpers.MockDatabaseScan(mock, ip)

		_, err := ObjectId(db, o)
		if err != nil {
			t.Fatal(err)
		}
	})
}
