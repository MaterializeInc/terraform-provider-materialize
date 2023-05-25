package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var inOwnership = map[string]interface{}{
	"object":      []interface{}{map[string]interface{}{"name": "table", "schema_name": "schema", "database_name": "database"}},
	"object_type": "TABLE",
	"role_name":   "my_role",
}

func TestResourceOwnershipDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Ownership().Schema, inOwnership)
	r.NotNil(d)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		if err := ownershipDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})

	if d.Id() != "" {
		t.Errorf("State id not set to empty string")
	}
}
