package resources

import (
	"context"
	"testing"
	"time"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestResourceIndexCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":         "index",
		"default":      false,
		"obj_name":     []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
		"col_expr":     []interface{}{map[string]interface{}{"field": "column"}},
		"cluster_name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT \(column\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_indexes.name = 'index' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, ip)

		// Query Params
		pp := `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_indexes.id = 'u1'`
		testhelpers.MockIndexColumnScan(mock, cp)

		if err := indexCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// Confirm id is updated with region for 0.4.0
func TestResourceIndexReadIdMigration(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "index",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_indexes.id = 'u1'`
		testhelpers.MockIndexColumnScan(mock, cp)

		if err := indexRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceIndexDelete(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":     "index",
		"default":  false,
		"obj_name": []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := indexDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceIndexDefaultCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name":         "index",
		"default":      true,
		"obj_name":     []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
		"cluster_name": "cluster",
	}
	d := schema.TestResourceDataRaw(t, Index().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE DEFAULT INDEX IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Update the WHERE clause to include the default index name pattern
		ip := `WHERE mz_databases.name = 'database' AND mz_indexes.name = 'source_primary_idx' AND mz_objects.name = 'source' AND mz_objects.type IN \('source', 'view', 'materialized-view'\) AND mz_schemas.name = 'schema'`
		testhelpers.MockIndexScan(mock, ip)

		// Query Params
		pp := `WHERE mz_indexes.id = 'u1' AND mz_objects.type IN \('source', 'view', 'materialized-view'\)`
		testhelpers.MockIndexScan(mock, pp)

		// Query Columns
		cp := `WHERE mz_indexes.id = 'u1'`
		testhelpers.MockIndexColumnScan(mock, cp)

		if err := indexCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func indexDeleteData(t *testing.T) *schema.ResourceData {
	in := map[string]interface{}{
		"name":     "index",
		"default":  false,
		"obj_name": []interface{}{map[string]interface{}{"name": "source", "schema_name": "schema", "database_name": "database"}},
	}
	return schema.TestResourceDataRaw(t, Index().Schema, in)
}

var dependentsError = &pgconn.PgError{Severity: "ERROR", Code: "2BP01", Message: `cannot drop index "index": still depended upon by materialized view "mv"`}

// terraform destroy can drop the reading materialized view in parallel; the
// drop succeeds once it is gone.
func TestResourceIndexDeleteRetriesWhileReadersAreDropped(t *testing.T) {
	r := require.New(t)
	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnError(dependentsError)
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		r.False(indexDelete(context.TODO(), indexDeleteData(t), db).HasError())
		r.NoError(mock.ExpectationsWereMet())
	})
}

// A reader that is not going away keeps the drop failing; the dependency error
// comes back once the retry window is over.
func TestResourceIndexDeleteGivesUpOnPersistentReaders(t *testing.T) {
	r := require.New(t)
	orig := indexDropRetryTimeout
	indexDropRetryTimeout = time.Second
	t.Cleanup(func() { indexDropRetryTimeout = orig })

	attempts := 0
	err := dropIndex(context.TODO(), func() error {
		attempts++
		return dependentsError
	})
	r.True(materialize.IsDependentObjectsStillExist(err))
	r.GreaterOrEqual(attempts, 2)
}

// Any other failure is returned straight away rather than retried.
func TestResourceIndexDeleteDoesNotRetryOtherErrors(t *testing.T) {
	r := require.New(t)
	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnError(&pgconn.PgError{Severity: "ERROR", Code: "XX000", Message: "internal error"})
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		diags := indexDelete(context.TODO(), indexDeleteData(t), db)
		r.True(diags.HasError())
		r.Contains(diags[0].Summary, "(SQLSTATE XX000)")
		r.Error(mock.ExpectationsWereMet(), "the second DROP must not have run")
	})
}
