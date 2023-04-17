package resources

var readSource = `
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
WHERE mz_sources.id = 'u1';`
