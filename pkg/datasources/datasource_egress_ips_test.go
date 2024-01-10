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

func TestEgressIpsDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, EgressIps().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		ir := mock.NewRows([]string{"egress_ip"}).
			AddRow("egress_ip")
		mock.ExpectQuery(`SELECT egress_ip FROM materialize.mz_catalog.mz_egress_ips;`).WillReturnRows(ir)

		if err := EgressIpsRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
