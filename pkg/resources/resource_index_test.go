package resources

import (
	"context"
	"terraform-materialize/pkg/testhelpers"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceIndexDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":     "index",
		"default":  false,
		"obj_name": []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := indexDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
