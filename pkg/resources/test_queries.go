package resources

import "github.com/DATA-DOG/go-sqlmock"

var readSinkParams = `
	SELECT
		mz_sinks.name,
		mz_schemas.name,
		mz_databases.name,
		mz_sinks.size,
		mz_connections.name AS connection_name,
		mz_clusters.name AS cluster_name
	FROM mz_sinks
	JOIN mz_schemas
		ON mz_sinks.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_connections
		ON mz_sinks.connection_id = mz_connections.id
	JOIN mz_clusters
		ON mz_sinks.cluster_id = mz_clusters.id
	WHERE mz_sinks.id = 'u1';
`

func mockSinkParams(mock sqlmock.Sqlmock) {
	ip := sqlmock.NewRows([]string{"name", "schema", "database", "size", "connection_name", "cluster_name"}).
		AddRow("sink", "schema", "database", "small", "conn", "cluster")
	mock.ExpectQuery(readSinkParams).WillReturnRows(ip)
}

var readSourceParams = `
	SELECT
		mz_sources.name,
		mz_schemas.name,
		mz_databases.name,
		mz_sources.size,
		mz_connections.name as connection_name,
		mz_clusters.name as cluster_name
	FROM mz_sources
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_connections
		ON mz_sources.connection_id = mz_connections.id
	JOIN mz_clusters
		ON mz_sources.cluster_id = mz_clusters.id
	WHERE mz_sources.id = 'u1';
`

func mockSourceParams(mock sqlmock.Sqlmock) {
	ip := sqlmock.NewRows([]string{"name", "schema", "database", "size", "connection_name", "cluster_name"}).
		AddRow("source", "schema", "database", "small", "conn", "cluster")
	mock.ExpectQuery(readSourceParams).WillReturnRows(ip)
}

var readConnectionParams = `
	SELECT
		mz_connections.name,
		mz_schemas.name,
		mz_databases.name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	WHERE mz_connections.id = 'u1';
`

func mockConnectionParams(mock sqlmock.Sqlmock) {
	ip := sqlmock.NewRows([]string{"name", "schema", "database"}).
		AddRow("conn", "schema", "database")
	mock.ExpectQuery(readConnectionParams).WillReturnRows(ip)
}

var readConnectionAwsPrivatelinkParams = `
	SELECT
		mz_connections.name,
		mz_schemas.name,
		mz_databases.name,
		mz_aws_privatelink_connections.principal
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_aws_privatelink_connections
		ON mz_connections.id = mz_aws_privatelink_connections.id
	WHERE mz_connections.id = 'u1';
`

func mockConnectionAwsPrivatelinkParams(mock sqlmock.Sqlmock) {
	ip := sqlmock.NewRows([]string{"name", "schema", "database", "principal"}).
		AddRow("conn", "schema", "database", "principal")
	mock.ExpectQuery(readConnectionAwsPrivatelinkParams).WillReturnRows(ip)
}

var readConnectionSshTunnelParams = `
	SELECT
		mz_connections.name,
		mz_schemas.name,
		mz_databases.name,
		mz_ssh_tunnel_connections.public_key_1,
		mz_ssh_tunnel_connections.public_key_2
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_ssh_tunnel_connections
		ON mz_connections.id = mz_ssh_tunnel_connections.id
	WHERE mz_connections.id = 'u1';
`

func mockConnectionSshTunnelParams(mock sqlmock.Sqlmock) {
	ip := sqlmock.NewRows([]string{"name", "schema", "database", "pk1", "pk2"}).
		AddRow("conn", "schema", "database", "pk1", "pk2")
	mock.ExpectQuery(readConnectionSshTunnelParams).WillReturnRows(ip)
}
