package testhelpers

import (
	"fmt"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

var defaultPrivilege = pq.StringArray{"s1=arwd/s1", "u1=UC/u18", "u8=arw/s1"}

func mockQueryBuilder(query, predicate, order string) string {
	q := strings.Builder{}
	q.WriteString(query)

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
	b := `
	SELECT
		mz_cluster_replicas.id,
		mz_cluster_replicas.name AS replica_name,
		mz_clusters.name AS cluster_name,
		mz_cluster_replicas.size,
		mz_cluster_replicas.availability_zone,
		mz_cluster_replicas.disk,
		comments.comment AS comment
	FROM mz_cluster_replicas
	JOIN mz_clusters
		ON mz_cluster_replicas.cluster_id = mz_clusters.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'cluster-replica'
	\) comments
		ON mz_cluster_replicas.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "replica_name", "cluster_name", "size", "availability_zone", "disk", "comment"}).
		AddRow("u1", "replica", "cluster", "small", "use1-az2", false, "comment")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockClusterScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_clusters.id,
		mz_clusters.name,
		mz_clusters.managed,
		mz_clusters.size,
		mz_clusters.replication_factor,
		mz_clusters.disk,
		mz_clusters.availability_zones,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_clusters.privileges
	FROM mz_clusters
	JOIN mz_roles
		ON mz_clusters.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'cluster'
	\) comments
		ON mz_clusters.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	az := pq.StringArray{"use1-az1", "use1-az2", "use1-az3"}
	ir := mock.NewRows([]string{"id", "name", "managed", "size", "replication_factor", "disk", "availability_zones", "comment", "owner_name", "privileges"}).
		AddRow("u1", "cluster", true, "small", 2, true, az, "comment", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_connections.type AS connection_type,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_connections.privileges
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	\) comments
		ON mz_connections.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "connection_type", "owner_name", "privileges"}).
		AddRow("u1", "connection", "schema", "database", "kafka", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionAwsScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_aws_connections.endpoint,
		mz_aws_connections.region AS aws_region,
		mz_aws_connections.access_key_id,
		mz_aws_connections.access_key_id_secret_id,
		mz_aws_connections.secret_access_key_secret_id,
		mz_aws_connections.session_token,
		mz_aws_connections.session_token_secret_id,
		mz_aws_connections.assume_role_arn,
		mz_aws_connections.assume_role_session_name,
		comments.comment AS comment,
		mz_aws_connections.principal,
		mz_aws_connections.external_id,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_aws_connections AS mz_aws_connections
		ON mz_connections.id = mz_aws_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	\) comments
		ON mz_connections.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
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
	b := `
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_aws_privatelink_connections.principal,
		comments.comment AS comment,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_aws_privatelink_connections
		ON mz_connections.id = mz_aws_privatelink_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	\) comments
		ON mz_connections.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "principal"}).
		AddRow("u1", "connection", "schema", "database", "principal")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockConnectionSshTunnelScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_ssh_tunnel_connections.public_key_1,
		mz_ssh_tunnel_connections.public_key_2,
		comments.comment AS comment,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_ssh_tunnel_connections
		ON mz_connections.id = mz_ssh_tunnel_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	\) comments
		ON mz_connections.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "public_key_1", "public_key_2"}).
		AddRow("u1", "connection", "schema", "database", "key_1", "key_2")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDefaultPrivilegeScan(mock sqlmock.Sqlmock, predicate, objectType string) {
	b := `
	SELECT
		mz_default_privileges.object_type,
		mz_default_privileges.grantee AS grantee_id,
		\(CASE WHEN mz_default_privileges.grantee = 'p' THEN 'PUBLIC' ELSE grantee.name END\) AS grantee_name,
		mz_default_privileges.role_id AS target_id,
		\(CASE WHEN mz_default_privileges.role_id = 'p' THEN 'PUBLIC' ELSE target.name END\) AS target_name,
		mz_default_privileges.database_id AS database_id,
		mz_default_privileges.schema_id AS schema_id,
		mz_default_privileges.privileges
	FROM mz_default_privileges
	LEFT JOIN mz_roles AS grantee
		ON mz_default_privileges.grantee = grantee.id
	LEFT JOIN mz_roles AS target
		ON mz_default_privileges.role_id = target.id
	LEFT JOIN mz_schemas
		ON mz_default_privileges.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_default_privileges.database_id = mz_databases.id`

	q := mockQueryBuilder(b, predicate, "")

	ir := mock.NewRows([]string{"object_type", "grantee_id", "grantee_name", "target_id", "target_name", "database_id", "schema_id", "privileges"}).
		AddRow(objectType, "u2", "grantee", "u4", "target", "u1", "u3", "U").
		AddRow(objectType, "u1", "grantee", "u1", "target", nil, nil, "Ur").
		AddRow(objectType, "u3", "grantee", "u6", "target", "u2", nil, "rw")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDatabaseScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_databases.id,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_databases.privileges
	FROM mz_databases
	JOIN mz_roles
		ON mz_databases.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'database'
	\) comments
		ON mz_databases.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockIndexColumnScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_columns.id,
		mz_columns.name,
		mz_columns.position,
		mz_columns.nullable,
		mz_columns.type,
		mz_columns.default,
		CASE WHEN mz_index_columns.index_id IS NOT NULL THEN true ELSE false END AS indexed_column,
		mz_indexes.name AS index_name,
		mz_indexes.id AS index_id
	FROM mz_columns
	LEFT JOIN mz_indexes
		ON mz_columns.id = mz_indexes.on_id
	LEFT JOIN mz_index_columns
		ON mz_index_columns.index_id = mz_indexes.id
		AND mz_index_columns.on_position = mz_columns.position`

	q := mockQueryBuilder(b, predicate, "ORDER BY mz_columns.position")
	ir := mock.NewRows([]string{"id", "name", "position", "nullable", "type", "default", "indexed_column", "index_name", "index_id"}).
		AddRow("u1", "column", "1", "true", "integer", "", "true", "index", "u1")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockIndexScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_indexes.id,
		mz_indexes.name AS index_name,
		mz_objects.name AS obj_name,
		mz_schemas.name AS obj_schema_name,
		mz_databases.name AS obj_database_name,
		comments.comment AS comment
	FROM mz_indexes
	JOIN mz_objects
		ON mz_indexes.on_id = mz_objects.id
	LEFT JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'index'
	\) comments
		ON mz_indexes.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "index_name", "obj_name", "obj_schema_name", "obj_database_name"}).
		AddRow("u1", "index", "obj", "schema", "database")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockMaterializeViewScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_materialized_views.id,
		mz_materialized_views.name AS materialized_view_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_clusters.name AS cluster_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_materialized_views.create_sql,
		mz_materialized_views.privileges
	FROM mz_materialized_views
	JOIN mz_schemas
		ON mz_materialized_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_clusters
		ON mz_materialized_views.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_materialized_views.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'materialized-view'
		AND object_sub_id IS NULL
	\) comments
		ON mz_materialized_views.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "materialized_view_name", "schema_name", "database_name", "cluster_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "cluster", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSystemPrivilege(mock sqlmock.Sqlmock) {
	b := "SELECT privileges FROM mz_system_privileges"

	ir := mock.NewRows([]string{"privileges"}).AddRow("s1=RBN/s1")
	mock.ExpectQuery(b).WillReturnRows(ir)
}

func MockRoleScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_roles.id,
		mz_roles.name AS role_name,
		mz_roles.inherit,
		comments.comment AS comment
	FROM mz_roles
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'role'
	\) comments
		ON mz_roles.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "role_name", "inherit"}).
		AddRow("u1", "joe", true)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockRoleGrantScan(mock sqlmock.Sqlmock) {
	q := `
	SELECT
		mz_role_members.role_id,
		mz_role_members.member,
		mz_role_members.grantor
	FROM mz_role_members`

	ir := mock.NewRows([]string{"role_id", "member", "grantor"}).
		AddRow("u2", "u3", "u3").
		AddRow("u1", "u1", "s1")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSchemaScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_schemas.id,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_schemas.privileges
	FROM mz_schemas JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_schemas.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'schema'
	\) comments
		ON mz_schemas.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSecretScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
		SELECT
		mz_secrets.id,
		mz_secrets.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,	
		mz_roles.name AS owner_name,
		mz_secrets.privileges
	FROM mz_secrets
	JOIN mz_schemas
		ON mz_secrets.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_secrets.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'secret'
	\) comments
		ON mz_secrets.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "secret", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSinkScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_sinks.id,
		mz_sinks.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sinks.type AS sink_type,
		COALESCE\(mz_sinks.size, mz_clusters.size\) AS size,
		mz_sinks.envelope_type,
		mz_connections.name as connection_name,
		mz_clusters.name as cluster_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name
	FROM mz_sinks
	JOIN mz_schemas
		ON mz_sinks.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_connections
		ON mz_sinks.connection_id = mz_connections.id
	LEFT JOIN mz_clusters
		ON mz_sinks.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_sinks.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'sink'
	\) comments
		ON mz_sinks.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "sink_type", "size", "envelope_type", "connection_name", "cluster_name", "owner_name"}).
		AddRow("u1", "sink", "schema", "database", "kafka", "small", "JSON", "conn", "cluster", "joe")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSourceScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_sources.id,
		mz_sources.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.type AS source_type,
		COALESCE\(mz_sources.size, mz_clusters.size\) AS size,
		mz_sources.envelope_type,
		mz_connections.name as connection_name,
		conn_schemas.name as connection_schema_name,
		conn_databases.name as connection_database_name,
		mz_clusters.name as cluster_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_webhook_sources.url AS webhook_url,
		mz_sources.privileges
	FROM mz_sources
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_connections
		ON mz_sources.connection_id = mz_connections.id
	LEFT JOIN mz_schemas conn_schemas
		ON mz_connections.schema_id = conn_schemas.id
	LEFT JOIN mz_databases conn_databases
		ON conn_schemas.database_id = conn_databases.id
	LEFT JOIN mz_clusters
		ON mz_sources.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_sources.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'source'
	\) comments
		ON mz_sources.id = comments.id
	LEFT JOIN mz_internal.mz_webhook_sources
		ON mz_sources.id = mz_webhook_sources.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "connection_schema_name", "connection_database_name", "cluster_name", "comment", "owner_name", "webhook_url", "privileges"}).
		AddRow("u1", "source", "schema", "database", "kafka", "small", "BYTES", "conn", "public", "materialize", "cluster", nil, "joe", "https://webhook.url/example", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
WITH dependencies AS \(
	SELECT
		mz_object_dependencies.object_id,
		mz_object_dependencies.referenced_object_id,
		mz_objects.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_objects.type,
		mz_object_dependencies.object_id AS filter_id,
		'object' AS source_type
	FROM mz_internal.mz_object_dependencies
	JOIN mz_objects
		ON mz_object_dependencies.referenced_object_id = mz_objects.id
	JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	UNION
	SELECT
		mz_object_dependencies.object_id,
		mz_object_dependencies.referenced_object_id,
		mz_objects.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_objects.type,
		mz_object_dependencies.referenced_object_id AS filter_id,
		'reference' AS source_type
	FROM mz_internal.mz_object_dependencies
	JOIN mz_objects
		ON mz_object_dependencies.object_id = mz_objects.id
	JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id\)
	SELECT \* FROM dependencies`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "filter_id", "source_type"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "u1", "object")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockPosgresSubsourceScan mocks the scan of a postgres source
func MockPosgresSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT DISTINCT
		mz_sources.id AS object_id,
		subsources.id AS referenced_object_id,
		mz_sources.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.type,
		mz_postgres_source_tables.table_name AS upstream_table_name,
		mz_postgres_source_tables.schema_name AS upstream_table_schema
	FROM mz_sources AS subsources
	JOIN mz_internal.mz_object_dependencies
		ON subsources.id = mz_object_dependencies.referenced_object_id
	JOIN mz_sources
		ON mz_sources.id = mz_object_dependencies.object_id
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_postgres_source_tables
		ON mz_sources.id = mz_postgres_source_tables.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "upstream_table_name", "upstream_table_schema"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "table", "table_schema")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockMysqlSubsourceScan mocks the scan of a postgres source
func MockMysqlSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT DISTINCT
		mz_sources.id AS object_id,
		subsources.id AS referenced_object_id,
		mz_sources.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.type,
		mz_mysql_source_tables.table_name AS upstream_table_name,
		mz_mysql_source_tables.schema_name AS upstream_table_schema
	FROM mz_sources AS subsources
	JOIN mz_internal.mz_object_dependencies
		ON subsources.id = mz_object_dependencies.referenced_object_id
	JOIN mz_sources
		ON mz_sources.id = mz_object_dependencies.object_id
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_mysql_source_tables
		ON mz_sources.id = mz_mysql_source_tables.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type", "upstream_table_name", "upstream_table_schema"}).
		AddRow("u1", "u2", "object", "schema", "database", "source", "table", "table_schema")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

// MockSQLServerSubsourceScan mocks the scan of a SQL Server source
func MockSQLServerSubsourceScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT DISTINCT
		mz_sources.id AS object_id,
		subsources.id AS referenced_object_id,
		mz_sources.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.type,
		mz_sql_server_source_tables.table_name AS upstream_table_name,
		mz_sql_server_source_tables.schema_name AS upstream_table_schema
	FROM mz_sources AS subsources
	JOIN mz_internal.mz_object_dependencies
		ON subsources.id = mz_object_dependencies.referenced_object_id
	JOIN mz_sources
		ON mz_sources.id = mz_object_dependencies.object_id
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_sql_server_source_tables
		ON mz_sources.id = mz_sql_server_source_tables.id`

	q := mockQueryBuilder(b, predicate, "")
	// TODO: Add back upstream_table_name and upstream_table_schema columns when mz_sqlserver_source_tables is implemented
	ir := mock.NewRows([]string{"object_id", "referenced_object_id", "object_name", "schema_name", "database_name", "type"}).
		AddRow("u1", "u2", "table1", "schema", "database", "subsource").
		AddRow("u1", "u2", "table2", "schema", "database", "subsource")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTableColumnScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_columns.id,
		mz_columns.name,
		mz_columns.position,
		mz_columns.nullable,
		comments.comment,
		mz_columns.type,
		mz_columns.default
	FROM mz_columns
	LEFT JOIN \(
		SELECT id, object_sub_id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
	\) comments
		ON mz_columns.id = comments.id
		AND mz_columns.position = comments.object_sub_id`

	q := mockQueryBuilder(b, predicate, "ORDER BY mz_columns.position")
	ir := mock.NewRows([]string{"id", "name", "position", "nullable", "type", "default"}).
		AddRow("u1", "column", "1", "true", "integer", "")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSystemGrantScan(mock sqlmock.Sqlmock) {
	q := `SELECT privileges FROM mz_system_privileges`
	ir := mock.NewRows([]string{"privileges"}).
		AddRow("u1=B/s1").AddRow("u9=RBN/s1").AddRow("u5=B/s1")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTableScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
		AND object_sub_id IS NULL
	\) comments
		ON mz_tables.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "comment", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "comment", "materialize", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTypeScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_types.id,
		mz_types.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_types.category,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_types.privileges
	FROM mz_types
	JOIN mz_schemas
		ON mz_types.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_types.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'type'
	\) comments
		ON mz_types.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "category", "owner_name", "privileges"}).
		AddRow("u1", "type", "schema", "database", "category", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockViewScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_views.id,
		mz_views.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_views.create_sql,
		mz_views.privileges
	FROM mz_views
	JOIN mz_schemas
		ON mz_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_views.owner_id = mz_roles.id
	LEFT JOIN \(
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'view'
		AND object_sub_id IS NULL
	\) comments
		ON mz_views.id = comments.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := sqlmock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "joe", defaultPrivilege)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockNetworkPolicyScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
		WITH policy AS \( 
			SELECT 
				mz_network_policies\.id, 
				mz_network_policies\.name AS policy_name, 
				comments\.comment AS comment, 
				mz_roles\.name AS owner_name, 
				mz_network_policies\.privileges 
			FROM mz_internal\.mz_network_policies 
			JOIN mz_roles 
				ON mz_network_policies\.owner_id = mz_roles\.id 
			LEFT JOIN \( 
				SELECT id, comment 
				FROM mz_internal\.mz_comments 
				WHERE object_type = 'network-policy' 
			\) comments 
				ON mz_network_policies\.id = comments\.id 
		\), 
		rules AS \( 
			SELECT 
				policy_id, 
				jsonb_agg\( 
					jsonb_build_object\( 
						'rule_name', name, 
						'rule_action', action, 
						'rule_direction', direction, 
						'rule_address', address 
					\) 
				\) as rules 
			FROM mz_internal\.mz_network_policy_rules 
			GROUP BY policy_id 
		\) 
		SELECT 
			policy\.\*, 
			COALESCE\(rules\.rules, '\[\]'::json\) as rules 
		FROM policy 
		LEFT JOIN rules 
			ON policy\.id = rules\.policy_id`

	q := mockQueryBuilder(b, predicate, "")
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
		pq.StringArray{"s1=U/s1"},
		`[{"rule_name":"minnesota","rule_action":"allow","rule_direction":"ingress","rule_address":"2.3.4.5/32"},
          {"rule_name":"new_york","rule_action":"allow","rule_direction":"ingress","rule_address":"1.2.3.4/28"}]`,
	)
	mock.ExpectQuery(q).WillReturnRows(ir)
}
