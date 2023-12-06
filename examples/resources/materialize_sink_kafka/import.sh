# Sinks can be imported using the sink id:
terraform import materialize_sink_kafka.example_sink_kafka <region>:<sink_id>

# Sink id and information be found in the `mz_catalog.mz_sinks` table
# The region is the region where the database is located (e.g. aws/us-east-1)
