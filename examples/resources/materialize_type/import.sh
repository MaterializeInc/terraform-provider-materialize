# Types can be imported using the type id:
terraform import materialize_type.example_type <region>:<type_id>

# Type id and information be found in the `mz_catalog.mz_types` table
# The region is the region where the database is located (e.g. aws/us-east-1)
```
