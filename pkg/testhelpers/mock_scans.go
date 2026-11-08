package testhelpers

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
)

var defaultPrivilege = materialize.StringArray{"s1=arwd/s1", "u1=UC/u18", "u8=arw/s1", "=UC/s1"}

// mockQueryBuilder builds the regex a mock expects from the query the provider
// runs, so a change to the query does not need a matching change here.
func mockQueryBuilder(query, predicate, order string) string {
	q := strings.Builder{}
	q.WriteString(regexp.QuoteMeta(query))

	if predicate != "" {
		q.WriteString(fmt.Sprintf(" %s", predicate))
	}

	if order != "" {
		q.WriteString(fmt.Sprintf(" %s", order))
	}

	q.WriteString(`;`)

	return q.String()
}

func MockClusterReplicaScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ClusterReplicaQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "replica_name", "cluster_name", "size", "availability_zone", "disk", "comment"}).
		AddRow("u1", "replica", "cluster", "small", "use1-az2", false, "comment")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func clusterScanQuery(predicate string) string {
	return mockQueryBuilder(materialize.ClusterQuery.Statement(), predicate, "")
}

var clusterScanColumns = []string{"id", "name", "managed", "size", "replication_factor", "disk", "availability_zones", "comment", "owner_name", "privileges"}

func MockClusterScan(mock sqlmock.Sqlmock, predicate string) {
	az := materialize.StringArray{"use1-az1", "use1-az2", "use1-az3"}
	ir := mock.NewRows(clusterScanColumns).
		AddRow("u1", "cluster", true, "small", 2, true, az, "comment", "joe", defaultPrivilege)
	mock.ExpectQuery(clusterScanQuery(predicate)).WillReturnRows(ir)
}

// MockClusterScanNoRows mocks a cluster scan that matches nothing, simulating a
// cluster that no longer exists.
func MockClusterScanNoRows(mock sqlmock.Sqlmock, predicate string) {
	mock.ExpectQuery(clusterScanQuery(predicate)).WillReturnRows(mock.NewRows(clusterScanColumns))
}

// MockClusterAutoScalingScan mocks the best-effort read of a cluster's
// autoscaling strategy. Pass an empty hydrationSize to simulate no configured
// strategy (no rows); lingerSeconds is the linger duration in seconds.
func MockClusterAutoScalingScan(mock sqlmock.Sqlmock, hydrationSize string, lingerSeconds int64) {
	q := `
		SELECT
			strategy->'on_hydration'->>'hydration_size' AS hydration_size,
			\(strategy->'on_hydration'->'linger_duration'->>'secs'\)::bigint AS linger_seconds
		FROM mz_internal.mz_cluster_auto_scaling_strategies
		WHERE cluster_id = \$1`

	rows := mock.NewRows([]string{"hydration_size", "linger_seconds"})
	if hydrationSize != "" {
		rows.AddRow(hydrationSize, lingerSeconds)
	}
	mock.ExpectQuery(q).WillReturnRows(rows)
}

// MockClusterReconfigurationScan mocks the read of an in-flight graceful resize.
// Pass an empty size to simulate no in-progress reconfiguration (no rows).
func MockClusterReconfigurationScan(mock sqlmock.Sqlmock, size string, replicationFactor int64) {
	q := `
		SELECT
			target->>'size' AS size,
			\(target->>'replication_factor'\)::bigint AS replication_factor
		FROM mz_internal.mz_cluster_reconfigurations
		WHERE cluster_id = .+ AND status = 'in-progress'`

	rows := mock.NewRows([]string{"size", "replication_factor"})
	if size != "" {
		rows.AddRow(size, replicationFactor)
	}
	mock.ExpectQuery(q).WillReturnRows(rows)
}

func MockConnectionScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ConnectionQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "connection_type", "owner_name", "privileges"}).
		AddRow("u1", "connection", "schema", "database", "kafka", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionAwsScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ConnectionAwsQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{
		"id",
		"connection_name",
		"schema_name",
		"database_name",
		"endpoint",
		"aws_region",
		"access_key_id",
		"access_key_id_secret_id",
		"secret_access_key_secret_id",
		"session_token",
		"session_token_secret_id",
		"assume_role_arn",
		"assume_role_session_name",
		"comment",
		"principal",
		"external_id",
		"owner_name",
	}).AddRow(
		"u1",
		"connection",
		"schema",
		"database",
		"localhost",
		"us-east-1",
		"foo",
		"u1",
		"u1",
		"bar",
		"u1",
		"arn:aws:iam::123456789012:user/JohnDoe",
		"s3-access-example",
		"comment",
		"principal",
		"mz_12345678-1234-1234-1234-123456789012_u123",
		"owner_name",
	)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionAwsPrivatelinkScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ConnectionAwsPrivatelinkQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "principal"}).
		AddRow("u1", "connection", "schema", "database", "principal")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionSshTunnelScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ConnectionSshTunnelQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "public_key_1", "public_key_2"}).
		AddRow("u1", "connection", "schema", "database", "key_1", "key_2")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDefaultPrivilegeScan(mock sqlmock.Sqlmock, predicate, objectType string) {
	q := mockQueryBuilder(materialize.DefaultPrivilegeQuery.Statement(), predicate, "")

	ir := mock.NewRows([]string{"object_type", "grantee_id", "grantee_name", "target_id", "target_name", "database_id", "schema_id", "privileges"}).
		AddRow(objectType, "u2", "grantee", "u4", "target", "u1", "u3", "U").
		AddRow(objectType, "u1", "grantee", "u1", "target", nil, nil, "Ur").
		AddRow(objectType, "u3", "grantee", "u6", "target", "u2", nil, "rw")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDatabaseScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.DatabaseQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockIndexColumnScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.IndexColumnQuery.Statement(), predicate, "ORDER BY mz_index_columns.index_position")
	ir := mock.NewRows([]string{"id", "name", "position", "nullable", "type", "default", "indexed_column", "index_name", "index_id"}).
		AddRow("u1", "column", "1", "true", "integer", "", "true", "index", "u1")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockIndexScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.IndexQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "index_name", "obj_name", "obj_schema_name", "obj_database_name"}).
		AddRow("u1", "index", "obj", "schema", "database")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockMaterializeViewScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.MaterializedViewQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "materialized_view_name", "schema_name", "database_name", "cluster_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "cluster", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSystemPrivilege(mock sqlmock.Sqlmock) {
	ir := mock.NewRows([]string{"privileges"}).AddRow("s1=RBN/s1")
	mock.ExpectQuery(regexp.QuoteMeta(materialize.SystemPrivilegeQuery)).WillReturnRows(ir)
}

func MockRoleScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.RoleQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "role_name", "inherit", "superuser", "login", "comment"}).
		AddRow("u1", "joe", true, false, false, "")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockRoleScanNoRows mocks a role lookup that finds no matching role, returning
// sql.ErrNoRows. Useful for exercising the create_if_not_exists path where the
// role does not yet exist in the catalog.
func MockRoleScanNoRows(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.RoleQuery.Statement(), predicate, "")
	mock.ExpectQuery(q).WillReturnError(sql.ErrNoRows)
}

func MockRoleGrantScan(mock sqlmock.Sqlmock) {
	ir := mock.NewRows([]string{"role_id", "member", "grantor"}).
		AddRow("u2", "u3", "u3").
		AddRow("u1", "u1", "s1")
	mock.ExpectQuery(regexp.QuoteMeta(materialize.RolePrivilegeQuery.Statement())).WillReturnRows(ir)
}

func MockSchemaScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SchemaQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSecretScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SecretQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "secret", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSinkScan(mock sqlmock.Sqlmock, predicate string) {
	mockSinkScanWithEnvelope(mock, predicate, "kafka", "JSON")
}

// MockSinkIcebergScan returns an Iceberg sink whose envelope_type carries the
// sink MODE, as mz_sinks reports it.
func MockSinkIcebergScan(mock sqlmock.Sqlmock, predicate, mode string) {
	mockSinkScanWithEnvelope(mock, predicate, "iceberg", mode)
}

func mockSinkScanWithEnvelope(mock sqlmock.Sqlmock, predicate, sinkType, envelope string) {
	q := mockQueryBuilder(materialize.SinkQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "sink_type", "size", "envelope_type", "connection_name", "cluster_name", "owner_name"}).
		AddRow("u1", "sink", "schema", "database", sinkType, "small", envelope, "conn", "cluster", "joe")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "connection_schema_name", "connection_database_name", "cluster_id", "cluster_name", "comment", "owner_name", "webhook_url", "privileges"}).
		AddRow("u1", "source", "schema", "database", "kafka", "small", "BYTES", "conn", "public", "materialize", "u5", "cluster", nil, "joe", "https://webhook.url/example", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTableMySQLScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTableMySQLQuery, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_name", "source_schema_name", "source_database_name", "upstream_table_name", "upstream_schema_name", "source_type", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "source", "public", "materialize", "upstream_table", "upstream_schema", "mysql", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTablePostgresScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTablePostgresQuery, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_name", "source_schema_name", "source_database_name", "upstream_table_name", "upstream_schema_name", "source_type", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "source", "public", "materialize", "upstream_table", "upstream_schema", "postgres", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTableKafkaScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTableKafkaQuery, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_name", "source_schema_name", "source_database_name", "upstream_table_name", "source_type", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "source", "public", "materialize", "topic", "kafka", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTableScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTableQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_name", "source_schema_name", "source_database_name", "upstream_table_name", "upstream_schema_name", "source_type", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "source", "public", "materialize", "table", "schema", "KAFKA", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTableWebhookScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTableWebhookQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceTableSQLServerScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceTableSQLServerQuery, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_name", "source_schema_name", "source_database_name", "upstream_table_name", "upstream_schema_name", "source_type", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "source", "public", "materialize", "upstream_table", "upstream_schema", "sqlserver", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceReferenceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SourceReferenceQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{
		"source_id", "namespace", "name", "updated_at", "columns",
		"source_name", "source_schema_name", "source_database_name", "source_type",
	}).AddRow(
		"source-id", "namespace", "reference_name", "2023-10-01T12:34:56Z",
		materialize.StringArray{"column1", "column2"},
		"source_name", "source_schema_name", "source_database_name", "source_type",
	)

	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.DependencyQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "filter_id", "source_type"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "u1", "object")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockPosgresSubsourceScan mocks the scan of a postgres source
func MockPosgresSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.PostgresSubsourceQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "upstream_table_name", "upstream_table_schema"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "table", "table_schema")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockMysqlSubsourceScan mocks the scan of a postgres source
func MockMysqlSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.MysqlSubsourceQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "upstream_table_name", "upstream_table_schema"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "table", "table_schema")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockSQLServerSubsourceScan mocks the scan of a SQL Server source
func MockSQLServerSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.SqlserverSubsourceQuery.Statement(), predicate, "")
	// TODO: Add back upstream_table_name and upstream_table_schema columns when mz_sqlserver_source_tables is implemented
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type"}).
		AddRow("u1", "u2", "table1", "schema", "database", "subsource").
		AddRow("u1", "u2", "table2", "schema", "database", "subsource")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTableColumnScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.TableColumnQuery.Statement(), predicate, "ORDER BY mz_columns.position")
	ir := mock.NewRows([]string{"id", "name", "position", "nullable", "type", "default"}).
		AddRow("u1", "column", "1", "true", "integer", "")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSystemGrantScan(mock sqlmock.Sqlmock) {
	ir := mock.NewRows([]string{"privileges"}).
		AddRow("u1=B/s1").AddRow("u9=RBN/s1").AddRow("u5=B/s1")
	mock.ExpectQuery(regexp.QuoteMeta(materialize.SystemPrivilegeQuery)).WillReturnRows(ir)
}

func MockTableScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.TableQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTypeScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.TypeQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "category", "owner_name", "privileges"}).
		AddRow("u1", "type", "schema", "database", "category", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockViewScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.ViewQuery.Statement(), predicate, "")
	ir := sqlmock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockNetworkPolicyScan(mock sqlmock.Sqlmock, predicate string) {
	q := mockQueryBuilder(materialize.NetworkPolicyQuery.Statement(), predicate, "")
	ir := mock.NewRows([]string{
		"id",
		"policy_name",
		"comment",
		"owner_name",
		"privileges",
		"rules",
	}).AddRow(
		"u1",
		"office_policy",
		"Network policy for office locations",
		"mz_system",
		materialize.StringArray{"s1=U/s1"},
		`[{"rule_name":"minnesota","rule_action":"allow","rule_direction":"ingress","rule_address":"2.3.4.5/32"},
          {"rule_name":"new_york","rule_action":"allow","rule_direction":"ingress","rule_address":"1.2.3.4/28"}]`,
	)
	mock.ExpectQuery(q).WillReturnRows(ir)
}
