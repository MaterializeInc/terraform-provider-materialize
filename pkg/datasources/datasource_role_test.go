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

func TestRoleDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{}
	d := schema.TestResourceDataRaw(t, Role().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "")

		if err := roleRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRoleDatasourceWithPattern(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"like_pattern": "prod_%",
	}
	d := schema.TestResourceDataRaw(t, Role().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "WHERE mz_roles.name LIKE 'prod_%'")

		if err := roleRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		roles := d.Get("roles").([]interface{})
		r.Equal(1, len(roles))
	})
}
