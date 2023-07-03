package testhelpers

import (
	"fmt"
	"strings"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

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
		mz_roles.name AS owner_name
	FROM mz_cluster_replicas
	JOIN mz_clusters
		ON mz_cluster_replicas.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_cluster_replicas.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "replica_name", "cluster_name", "size", "availability_zone", "owner_name"}).
		AddRow("u1", "replica", "cluster", "small", "use1-az2", "joe")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockClusterScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_clusters.id,
		mz_clusters.name,
		mz_roles.name AS owner_name,
		mz_clusters.privileges
	FROM mz_clusters
	JOIN mz_roles
		ON mz_clusters.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "owner_name", "privileges"}).
		AddRow("u1", "cluster", "joe", "{u1=UC/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockColumnScan(mock sqlmock.Sqlmock, predicate string) {
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
		AND mz_index_columns.index_position = mz_columns.position`

	q := mockQueryBuilder(b, predicate, "ORDER BY mz_columns.position")
	ir := mock.NewRows([]string{"id", "name", "position", "nullable", "type", "default", "indexed_column", "index_name", "index_id"}).
		AddRow("u1", "column", "1", "true", "integer", "", "false", "", "")
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
		mz_roles.name AS owner_name,
		mz_connections.privileges
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "connection_type", "owner_name", "privileges"}).
		AddRow("u1", "connection", "schema", "database", "kafka", "joe", "{u1=U/u18}")
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
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_aws_privatelink_connections
		ON mz_connections.id = mz_aws_privatelink_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id`

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
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_ssh_tunnel_connections
		ON mz_connections.id = mz_ssh_tunnel_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "connection_name", "schema_name", "database_name", "public_key_1", "public_key_2"}).
		AddRow("u1", "connection", "schema", "database", "key_1", "key_2")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDefaultPrivilegeScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_default_privileges.grantee AS grantee_id,
		mz_default_privileges.role_id,
		mz_default_privileges.schema_id AS schema_id,
		mz_default_privileges.database_id AS database_id,
		mz_default_privileges.object_type,
		mz_default_privileges.privileges
	FROM mz_default_privileges
	LEFT JOIN mz_schemas
		ON mz_default_privileges.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_default_privileges.database_id = mz_databases.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"grantee_id", "role_id", "schema_id", "database_id", "object_type", "privileges"}).
		AddRow("u1", "u1", "u1", "u1", "TABLES", "{u1=UC/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockDatabaseScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_databases.id,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_databases.privileges
	FROM mz_databases
	JOIN mz_roles
		ON mz_databases.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "database", "joe", "{u1=UC/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockIndexScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_indexes.id,
		mz_indexes.name AS index_name,
		mz_objects.name AS obj_name,
		mz_schemas.name AS obj_schema_name,
		mz_databases.name AS obj_database_name
	FROM mz_indexes
	JOIN mz_objects
		ON mz_indexes.on_id = mz_objects.id
	LEFT JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "index_name", "obj_name", "obj_schema_name", "obj_database_name"}).
		AddRow("u1", "index", "obj", "schema", "database")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockMaterailizeViewScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_materialized_views.id,
		mz_materialized_views.name AS materialized_view_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_clusters.name AS cluster_name,
		mz_roles.name AS owner_name,
		mz_materialized_views.privileges
	FROM mz_materialized_views
	JOIN mz_schemas
		ON mz_materialized_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_clusters
		ON mz_materialized_views.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_materialized_views.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "materialized_view_name", "schema_name", "database_name", "cluster_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "cluster", "joe", "{u1=r/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockRoleScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		id,
		name AS role_name,
		inherit,
		create_role,
		create_db,
		create_cluster
	FROM mz_roles`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "role_name", "inherit", "create_role", "create_db", "create_cluster"}).
		AddRow("u1", "joe", true, true, true, true)
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSchemaScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_schemas.id,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_schemas.privileges
	FROM mz_schemas JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_schemas.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "schema", "database", "joe", "{u1=UC/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockSecretScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
		SELECT
		mz_secrets.id,
		mz_secrets.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_secrets.privileges
	FROM mz_secrets
	JOIN mz_schemas
		ON mz_secrets.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_secrets.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "secret", "schema", "database", "joe", "{u1=U/u18}")
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
		mz_sinks.size,
		mz_sinks.envelope_type,
		mz_connections.name as connection_name,
		mz_clusters.name as cluster_name,
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
		ON mz_sinks.owner_id = mz_roles.id`

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
		mz_sources.size,
		mz_sources.envelope_type,
		mz_connections.name as connection_name,
		mz_clusters.name as cluster_name,
		mz_roles.name AS owner_name,
		mz_sources.privileges
	FROM mz_sources
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_connections
		ON mz_sources.connection_id = mz_connections.id
	LEFT JOIN mz_clusters
		ON mz_sources.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_sources.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "source_type", "size", "envelope_type", "connection_name", "cluster_name", "owner_name", "privileges"}).
		AddRow("u1", "source", "schema", "database", "kafka", "small", "BYTES", "conn", "cluster", "joe", "{u1=r/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockTableScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "table", "schema", "database", "materialize", "{u1=arwd/u18}")
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
		mz_roles.name AS owner_name,
		mz_types.privileges
	FROM mz_types
	JOIN mz_schemas
		ON mz_types.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_types.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := mock.NewRows([]string{"id", "name", "schema_name", "database_name", "category", "owner_name", "privileges"}).
		AddRow("u1", "type", "schema", "database", "category", "joe", "{u1=U/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}

func MockViewScan(mock sqlmock.Sqlmock, predicate string) {
	b := `
	SELECT
		mz_views.id,
		mz_views.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_views.privileges
	FROM mz_views
	JOIN mz_schemas
		ON mz_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_views.owner_id = mz_roles.id`

	q := mockQueryBuilder(b, predicate, "")
	ir := sqlmock.NewRows([]string{"id", "name", "schema_name", "database_name", "owner_name", "privileges"}).
		AddRow("u1", "view", "schema", "database", "joe", "{u1=r/u18}")
	mock.ExpectQuery(q).WillReturnRows(ir)
}
