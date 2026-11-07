package resources

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

const subsourcePredicate = `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`

type sourceReadCase struct {
	name       string
	resource   *schema.Resource
	in         map[string]interface{}
	read       func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	subsources func(sqlmock.Sqlmock, string)
	extra      map[string]interface{}
}

func sourceReadCases() []sourceReadCase {
	connection := []interface{}{map[string]interface{}{"name": "conn", "schema_name": "public", "database_name": "materialize"}}
	upstreamTable := []interface{}{map[string]interface{}{
		"upstream_name": "table", "upstream_schema_name": "table_schema",
		"name": "object", "schema_name": "schema", "database_name": "database",
	}}
	return []sourceReadCase{
		{
			name: "kafka", resource: SourceKafka(), in: inSourceKafka(), read: sourceKafkaRead,
			extra: map[string]interface{}{"kafka_connection": connection},
		},
		{
			name: "postgres", resource: SourcePostgres(), in: inSourcePostgresTable(), read: sourcePostgresRead,
			subsources: testhelpers.MockPosgresSubsourceScan,
			extra:      map[string]interface{}{"postgres_connection": connection, "table": upstreamTable},
		},
		{
			name: "mysql", resource: SourceMySQL(), in: inSourceMySQLTable(), read: sourceMySQLRead,
			subsources: testhelpers.MockMysqlSubsourceScan,
			extra:      map[string]interface{}{"mysql_connection": connection, "table": upstreamTable},
		},
		{
			name: "sqlserver", resource: SourceSQLServer(), in: inSourceSQLServer(), read: sourceSQLServerRead,
			subsources: testhelpers.MockSQLServerSubsourceScan,
			extra: map[string]interface{}{"sqlserver_connection": connection, "table": []interface{}{
				map[string]interface{}{"upstream_name": "", "upstream_schema_name": "", "name": "table1", "schema_name": "schema", "database_name": "database"},
				map[string]interface{}{"upstream_name": "", "upstream_schema_name": "", "name": "table2", "schema_name": "schema", "database_name": "database"},
			}},
		},
		{
			name: "load generator", resource: SourceLoadgen(), in: inSourceLoadgen(), read: sourceRead,
		},
		{
			name: "webhook", resource: SourceWebhook(), in: inSourceWebhook(), read: sourceRead,
			extra: map[string]interface{}{"url": "https://webhook.url/example"},
		},
	}
}

func stateValue(v interface{}) interface{} {
	if s, ok := v.(*schema.Set); ok {
		return s.List()
	}
	return v
}

// Every source type reads the same shared fields from the source row, plus the
// fields of its own type. The state starts out different from the row, and an
// id without a region gets one.
func TestResourceSourceReadState(t *testing.T) {
	for _, c := range sourceReadCases() {
		for _, id := range []string{"aws/us-east-1:u1", "u1"} {
			t.Run(c.name+"/"+id, func(t *testing.T) {
				r := require.New(t)
				in := c.in
				for _, k := range []string{"name", "schema_name", "database_name", "cluster_name", "ownership_role", "comment"} {
					in[k] = "stale"
				}
				d := schema.TestResourceDataRaw(t, c.resource.Schema, in)
				d.SetId(id)

				testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
					testhelpers.MockSourceScan(mock, `WHERE mz_sources.id = 'u1'`)
					if c.subsources != nil {
						c.subsources(mock, subsourcePredicate)
					}

					r.False(c.read(context.TODO(), d, db).HasError())
					r.NoError(mock.ExpectationsWereMet())
				})

				r.Equal("aws/us-east-1:u1", d.Id())
				assertSourceReadState(r, d, c.extra)
			})
		}
	}
}

func assertSourceReadState(r *require.Assertions, d *schema.ResourceData, extra map[string]interface{}) {
	want := map[string]interface{}{
		"name":               "source",
		"schema_name":        "schema",
		"database_name":      "database",
		"size":               "small",
		"cluster_name":       "cluster",
		"ownership_role":     "joe",
		"qualified_sql_name": `"database"."schema"."source"`,
		"comment":            "",
	}
	for k, v := range extra {
		want[k] = v
	}
	for k, v := range want {
		if list, ok := v.([]interface{}); ok && len(list) > 1 {
			r.ElementsMatch(list, stateValue(d.Get(k)), k)
			continue
		}
		r.Equal(v, stateValue(d.Get(k)), k)
	}
}

// A source that no longer exists is removed from the state.
func TestResourceSourceReadGone(t *testing.T) {
	for _, c := range sourceReadCases() {
		t.Run(c.name, func(t *testing.T) {
			r := require.New(t)
			d := schema.TestResourceDataRaw(t, c.resource.Schema, c.in)
			d.SetId("aws/us-east-1:u1")

			testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrNoRows)

				r.False(c.read(context.TODO(), d, db).HasError())
				r.NoError(mock.ExpectationsWereMet())
			})

			r.Equal("", d.Id())
		})
	}
}

// While a resize of the source's cluster is in flight, every source type
// reports the target size instead of the size the cluster is leaving.
func TestResourceSourceReadInFlightResize(t *testing.T) {
	for _, c := range sourceReadCases() {
		for _, pending := range []struct{ target, want string }{{"", "small"}, {"medium", "medium"}} {
			t.Run(c.name+"/"+pending.want, func(t *testing.T) {
				r := require.New(t)
				d := schema.TestResourceDataRaw(t, c.resource.Schema, c.in)
				d.SetId("aws/us-east-1:u1")

				testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
					testhelpers.MockSourceScanInCluster(mock, `WHERE mz_sources.id = 'u1'`, "u5")
					rows := mock.NewRows([]string{"size", "replication_factor"})
					if pending.target != "" {
						rows.AddRow(pending.target, 1)
					}
					mock.ExpectQuery(`FROM mz_internal.mz_cluster_reconfigurations`).WithArgs("u5").WillReturnRows(rows)
					if c.subsources != nil {
						c.subsources(mock, subsourcePredicate)
					}

					r.False(c.read(context.TODO(), d, db).HasError())
					r.NoError(mock.ExpectationsWereMet())
				})

				r.Equal(pending.want, d.Get("size"))
			})
		}
	}
}

// A failed resize lookup fails the read instead of writing a guessed size.
func TestResourceSourceReadResizeLookupError(t *testing.T) {
	for _, c := range sourceReadCases() {
		t.Run(c.name, func(t *testing.T) {
			r := require.New(t)
			d := schema.TestResourceDataRaw(t, c.resource.Schema, c.in)
			d.SetId("aws/us-east-1:u1")

			testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
				testhelpers.MockSourceScanInCluster(mock, `WHERE mz_sources.id = 'u1'`, "u5")
				mock.ExpectQuery(`FROM mz_internal.mz_cluster_reconfigurations`).WithArgs("u5").WillReturnError(errors.New("connection reset"))

				r.True(c.read(context.TODO(), d, db).HasError())
				r.NoError(mock.ExpectationsWereMet())
			})

			r.Equal("aws/us-east-1:u1", d.Id())
		})
	}
}
