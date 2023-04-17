package resources

var readConnection string = `
SELECT
	mz_connections.name,
	mz_schemas.name,
	mz_databases.name
FROM mz_connections
JOIN mz_schemas
	ON mz_connections.schema_id = mz_schemas.id
JOIN mz_databases
	ON mz_schemas.database_id = mz_databases.id
WHERE mz_connections.id = 'u1';`
