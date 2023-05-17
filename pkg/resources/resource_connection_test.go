package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestResourceConnectionRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, ConnectionKafka().Schema, inKafka)
	r.NotNil(d)
	d.SetId("u1")

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mockConnectionParams(mock)

		if err := connectionRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Ensure parameters set
		expectedParams := map[string]string{
			"name":               "conn",
			"schema_name":        "schema",
			"database_name":      "database",
			"qualified_sql_name": `"database"."schema"."conn"`,
		}

		for key, value := range expectedParams {
			v := d.Get(key).(string)
			if v != value {
				t.Fatalf("Unexpected parameter set for %s. Recieved: %s", key, v)
			}
		}
	})
}
