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

func TestCurrentDatabaseDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, CurrentDatabase().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"database"}).AddRow("materialize")
		mock.ExpectQuery(`SHOW DATABASE;`).WillReturnRows(ir)

		if err := currentDatabaseRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
