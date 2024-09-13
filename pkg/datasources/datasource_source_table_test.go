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

func TestSourceTableDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"schema_name":   "schema",
		"database_name": "database",
	}
	d := schema.TestResourceDataRaw(t, SourceTable().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		p := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSourceTableScan(mock, p)

		if err := sourceTableRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Verify the results
		tables := d.Get("tables").([]interface{})
		r.Equal(1, len(tables))

		table := tables[0].(map[string]interface{})
		r.Equal("u1", table["id"])
		r.Equal("table", table["name"])
		r.Equal("schema", table["schema_name"])
		r.Equal("database", table["database_name"])
		r.Equal("KAFKA", table["source_type"])
		// TODO: Update once upstream_name and upstream_schema_name are supported
		r.Equal("", table["upstream_name"])
		r.Equal("", table["upstream_schema_name"])
		r.Equal("comment", table["comment"])
		r.Equal("materialize", table["owner_name"])

		source := table["source"].([]interface{})[0].(map[string]interface{})
		r.Equal("source", source["name"])
		r.Equal("public", source["schema_name"])
		r.Equal("materialize", source["database_name"])
	})
}
