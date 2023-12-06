# Databases can be imported using the database id:
terraform import materialize_database.example_database <region>:<database_id>

# Database id and information be found in the `mz_catalog.mz_databases` table
# The region is the region where the database is located (e.g. aws/us-east-1)
