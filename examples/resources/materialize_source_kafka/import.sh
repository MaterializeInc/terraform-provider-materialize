# Sources can be imported using the source id:
terraform import materialize_source_kafka.example_source_kafka <region>:<source_id>

# Source id and information be found in the `mz_catalog.mz_sources` table
# The region is the region where the database is located (e.g. aws/us-east-1)
