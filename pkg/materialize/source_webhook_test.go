package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceWebhook = ObjectSchemaStruct{Name: "webhook_source", SchemaName: "schema", DatabaseName: "database"}
var checkOptions = []CheckOptionsStruct{
	{
		Field: "BODY",
		Alias: "bytes",
	},
	{
		Field: "HEADERS",
		Alias: "headers",
	},
}

func TestSourceWebhookCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADERS CHECK \( WITH \(BODY AS bytes\, HEADERS AS headers\) check_expression\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceWebhookBuilder(db, sourceWebhook)
		b.ClusterName("cluster")
		b.BodyFormat("JSON")
		b.IncludeHeaders(true)
		b.CheckOptions(checkOptions)
		b.CheckExpression("check_expression")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}
