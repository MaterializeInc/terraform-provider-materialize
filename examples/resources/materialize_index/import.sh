# Indexes can be imported using the index id:
terraform import materialize_index.example_index <region>:<index_id>

# Index id and information be found in the `mz_catalog.mz_indexes` table
# The region is the region where the database is located (e.g. aws/us-east-1)
