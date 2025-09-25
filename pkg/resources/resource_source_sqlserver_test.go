package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inSourceSQLServer = map[string]interface{}{
	"name":          "source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "test_cluster",
	"sqlserver_connection": []interface{}{
		map[string]interface{}{
			"name":          "sqlserver_connection",
			"schema_name":   "schema",
			"database_name": "database",
		},
	},
	"text_columns":    []interface{}{"dbo.table1.xml_column", "custom.table2.ntext_column"},
	"exclude_columns": []interface{}{"dbo.table1.geometry_column", "custom.table2.geography_column"},
	"table": []interface{}{
		map[string]interface{}{
			"upstream_name":        "table1",
			"upstream_schema_name": "dbo",
			"name":                 "renamed_table1",
		},
		map[string]interface{}{
			"upstream_name":        "table2",
			"upstream_schema_name": "custom",
		},
	},
	"expose_progress": []interface{}{
		map[string]interface{}{
			"name":          "progress",
			"schema_name":   "schema",
			"database_name": "database",
		},
	},
	"comment": "SQL Server source comment",
}

func TestResourceSourceSQLServerCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, inSourceSQLServer)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."source" IN CLUSTER "test_cluster" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(TEXT COLUMNS \(dbo.table1.xml_column, custom.table2.ntext_column\), EXCLUDE COLUMNS \(dbo.table1.geometry_column, custom.table2.geography_column\)\) FOR TABLES \("dbo"."table1" AS "database"."schema"."renamed_table1", "custom"."table2" AS "database"."schema"."table2"\) EXPOSE PROGRESS AS "database"."schema"."progress";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment
		mock.ExpectExec(`COMMENT ON SOURCE "database"."schema"."source" IS 'SQL Server source comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerCreateAllTablesMinimal(t *testing.T) {
	minimalInput := map[string]interface{}{
		"name":          "minimal_source",
		"schema_name":   "schema",
		"database_name": "database",
		"sqlserver_connection": []interface{}{
			map[string]interface{}{
				"name":          "sqlserver_connection",
				"schema_name":   "schema",
				"database_name": "database",
			},
		},
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, minimalInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with minimal configuration (FOR ALL TABLES)
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."minimal_source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'minimal_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerCreateWithTextColumnsOnly(t *testing.T) {
	textColumnsInput := map[string]interface{}{
		"name":          "text_cols_source",
		"schema_name":   "schema",
		"database_name": "database",
		"sqlserver_connection": []interface{}{
			map[string]interface{}{
				"name":          "sqlserver_connection",
				"schema_name":   "schema",
				"database_name": "database",
			},
		},
		"text_columns": []interface{}{"xml_data", "ntext_data"},
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, textColumnsInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with TEXT COLUMNS only
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."text_cols_source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(TEXT COLUMNS \(xml_data, ntext_data\)\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'text_cols_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerCreateWithSSL(t *testing.T) {
	sslInput := map[string]interface{}{
		"name":          "ssl_source",
		"schema_name":   "schema",
		"database_name": "database",
		"sqlserver_connection": []interface{}{
			map[string]interface{}{
				"name":          "sqlserver_connection",
				"schema_name":   "schema",
				"database_name": "database",
			},
		},
		"ssl_mode": "require",
		"ssl_certificate_authority": []interface{}{
			map[string]interface{}{
				"text": "-----BEGIN CERTIFICATE-----",
			},
		},
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, sslInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with SSL
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."ssl_source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(SSL MODE 'require', SSL CERTIFICATE AUTHORITY '-----BEGIN CERTIFICATE-----'\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'ssl_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerCreateWithSSLSecret(t *testing.T) {
	sslSecretInput := map[string]interface{}{
		"name":          "ssl_secret_source",
		"schema_name":   "schema",
		"database_name": "database",
		"sqlserver_connection": []interface{}{
			map[string]interface{}{
				"name":          "sqlserver_connection",
				"schema_name":   "schema",
				"database_name": "database",
			},
		},
		"ssl_mode": "verify-ca",
		"ssl_certificate_authority": []interface{}{
			map[string]interface{}{
				"secret": []interface{}{map[string]interface{}{
					"name":          "ssl_ca_secret",
					"schema_name":   "schema",
					"database_name": "database",
				}},
			},
		},
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, sslSecretInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with SSL secret
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."ssl_secret_source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(SSL MODE 'verify-ca', SSL CERTIFICATE AUTHORITY SECRET "database"."schema"."ssl_ca_secret"\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'ssl_secret_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerCreateWithExcludeColumnsOnly(t *testing.T) {
	excludeColumnsInput := map[string]interface{}{
		"name":          "exclude_cols_source",
		"schema_name":   "schema",
		"database_name": "database",
		"sqlserver_connection": []interface{}{
			map[string]interface{}{
				"name":          "sqlserver_connection",
				"schema_name":   "schema",
				"database_name": "database",
			},
		},
		"exclude_columns": []interface{}{"geometry_data", "geography_data"},
	}

	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, excludeColumnsInput)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create with EXCLUDE COLUMNS only
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."exclude_cols_source" FROM SQL SERVER CONNECTION "database"."schema"."sqlserver_connection" \(EXCLUDE COLUMNS \(geometry_data, geography_data\)\) FOR ALL TABLES;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'exclude_cols_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, inSourceSQLServer)
	d.SetId("aws/us-east-1:u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		p := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, p)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		if err := sourceSQLServerRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("source", d.Get("name"))
		r.Equal("schema", d.Get("schema_name"))
		r.Equal("database", d.Get("database_name"))
	})
}

func TestResourceSourceSQLServerUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, inSourceSQLServer)

	// Set current state
	d.SetId("u1")
	d.Set("name", "source")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Name change - unit tests always see empty string as old name
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."" RENAME TO "source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Add subsources (tables) - detected as changes in unit test
		mock.ExpectExec(`ALTER SOURCE "database"."schema"."source" ADD SUBSOURCE "dbo"."table1" AS "database"."schema"."renamed_table1", "custom"."table2" WITH \(TEXT COLUMNS \[dbo.table1.xml_column, custom.table2.ntext_column\]\);`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Comment change
		mock.ExpectExec(`COMMENT ON SOURCE "database"."schema"."source" IS 'SQL Server source comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		p := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, p)

		// Query Subsources - SQL Server
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_sources.type = 'subsource'`
		testhelpers.MockSQLServerSubsourceScan(mock, ps)

		// Execute the update function
		if err := sourceSQLServerUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceSQLServerDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceSQLServer().Schema, inSourceSQLServer)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP SOURCE "database"."schema"."source";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
