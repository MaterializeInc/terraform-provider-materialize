package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestResourceIndexReadPreservesIndexKeyOrder(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Index().Schema, map[string]interface{}{"name": "index"})
	d.SetId("u1")
	testhelpers.WithMockProviderMeta(t, func(meta *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		testhelpers.MockIndexScan(mock, `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`)
		// An index on (b, a) must follow index_position, not the table's (a, b) positions.
		rows := sqlmock.NewRows([]string{"id", "name", "position", "nullable", "type", "default", "indexed_column", "index_name", "index_id"}).
			AddRow("u2", "b", "2", true, "integer", "", true, "index", "u1").
			AddRow("u2", "a", "1", true, "integer", "", true, "index", "u1")
		mock.ExpectQuery(`(?s)SELECT.*FROM mz_columns.*WHERE mz_indexes.id = 'u1' ORDER BY mz_index_columns.index_position;`).WillReturnRows(rows)
		require.Empty(t, indexRead(context.Background(), d, meta))
		require.Equal(t, []interface{}{
			map[string]interface{}{"field": "b"},
			map[string]interface{}{"field": "a"},
		}, d.Get("col_expr"))
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
