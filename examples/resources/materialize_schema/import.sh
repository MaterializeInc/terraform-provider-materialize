# Schemas can be imported using the schema id:
terraform import materialize_schema.example_schema <role>:<schema_id>

# Schema id and information be found in the `mz_catalog.mz_schemas` table
# The role is the role where the database is located (e.g. aws/us-east-1)
