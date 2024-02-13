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

func TestSystemParametersDatasource(t *testing.T) {
	r := require.New(t)

	// No input required for SystemParameter data source
	d := schema.TestResourceDataRaw(t, SystemParameter().Schema, make(map[string]interface{}))
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Mock the SQL query response
		rows := sqlmock.NewRows([]string{"name", "setting", "description"}).
			AddRow("cluster", "quickstart", "Sets the current cluster (Materialize).")

		// Expect a query to be made and return the mocked rows
		mock.ExpectQuery("SHOW ALL;").WillReturnRows(rows)

		// Execute the data source read function
		if err := systemParameterRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Verify the results
		parameters, _ := d.Get("parameters").([]interface{})
		r.Len(parameters, 1) // Expecting a single parameter result

		// Perform more detailed checks on the result if necessary
		result := parameters[0].(map[string]interface{})
		r.Equal("cluster", result["name"])
		r.Equal("quickstart", result["setting"])
		r.Equal("Sets the current cluster (Materialize).", result["description"])
	})

}
