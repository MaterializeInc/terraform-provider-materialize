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

func TestCurrentClusterDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, CurrentCluster().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"cluster"}).AddRow("quickstart")
		mock.ExpectQuery(`SHOW CLUSTER;`).WillReturnRows(ir)

		if err := currentClusterRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
