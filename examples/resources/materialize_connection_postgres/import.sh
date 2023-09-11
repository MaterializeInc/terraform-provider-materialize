# Connections can be imported using the `connection id`:
terraform import materialize_connection_postgres.example <connection_id>

# You can find the `connection_id` from the following query:
SELECT id, name, type FROM mz_catalog.mz_connections;
SELECT id, name, type FROM mz_objects WHERE type = 'connection';