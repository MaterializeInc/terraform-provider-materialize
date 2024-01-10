package datasources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestIndexDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		p := `
		WHERE mz_databases.name = 'database'
		AND mz_objects.type IN \('source', 'view', 'materialized-view'\)
		AND mz_schemas.name = 'schema'`
		testhelpers.MockIndexScan(mock, p)

		if err := indexRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
