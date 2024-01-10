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

func TestSchemaDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, Schema().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		p := `WHERE mz_databases.name = 'database'`
		testhelpers.MockSchemaScan(mock, p)

		if err := schemaRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

}
