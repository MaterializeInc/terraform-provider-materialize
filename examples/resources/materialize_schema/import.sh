# Schemas can be imported using the schema id or name
terraform import materialize_schema.example_schema <region>:id:<schema_id>

# To import using the schema name, set the `identify_by_name` attribute to true
terraform import materialize_schema.example_schema <region>:name:<database>|<schema>

# Schema id and information can be found in the mz_catalog.mz_schemas table
# The region is the region where the database is located (e.g. aws/us-east-1)
