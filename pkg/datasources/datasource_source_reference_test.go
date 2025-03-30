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

func TestSourceReferenceDatasource(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"source_id": "source-id",
	}
	d := schema.TestResourceDataRaw(t, SourceReference().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		predicate := `WHERE sr.source_id = 'source-id'`
		testhelpers.MockSourceReferenceScan(mock, predicate)

		if err := sourceReferenceRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		// Verify the results
		references := d.Get("references").([]interface{})
		r.Equal(1, len(references))

		reference := references[0].(map[string]interface{})
		r.Equal("namespace", reference["namespace"])
		r.Equal("reference_name", reference["name"])
		r.Equal("2023-10-01T12:34:56Z", reference["updated_at"])
		r.Equal([]interface{}{"column1", "column2"}, reference["columns"])
		r.Equal("source_name", reference["source_name"])
		r.Equal("source_schema_name", reference["source_schema_name"])
		r.Equal("source_database_name", reference["source_database_name"])
		r.Equal("source_type", reference["source_type"])
	})
}
