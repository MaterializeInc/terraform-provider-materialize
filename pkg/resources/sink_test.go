package resources

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
WHERE mz_sinks.id = 'u1';`