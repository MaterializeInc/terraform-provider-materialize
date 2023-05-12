package resources

import "github.com/DATA-DOG/go-sqlmock"

var readSink = `
	SELECT
		mz_sinks.name,
		mz_schemas.name,
		mz_databases.name,
		mz_sinks.size,
		mz_connections.name as connection_name,
		mz_clusters.name as cluster_name
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
		AddRow("conn", "schema", "database", "small", "conn", "cluster")
	mock.ExpectQuery(readSourceParams).WillReturnRows(ip)
}
