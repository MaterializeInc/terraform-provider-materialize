# Source tables can be imported using the source table id:
terraform import materialize_source_table_kafka.example_source_table_kafka <region>:<source_table_id>

# Source id and information be found in the `mz_catalog.mz_tables` table
# The region is the region where the database is located (e.g. aws/us-east-1)
