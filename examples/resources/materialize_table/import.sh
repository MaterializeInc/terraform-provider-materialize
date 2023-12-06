# Tables can be imported using the table id:
terraform import materialize_table.example_table <region>:<table_id>

# Table id and information be found in the `mz_catalog.mz_tables` table
# The region is the region where the database is located (e.g. aws/us-east-1)
