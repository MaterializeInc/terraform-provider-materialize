# Connections can be imported using the connection id:
terraform import materialize_connection_postgres.example <region>:<connection_id>

# Connection id and information be found in the `mz_catalog.mz_connections` table
# The region is the region where the database is located (e.g. aws/us-east-1)
