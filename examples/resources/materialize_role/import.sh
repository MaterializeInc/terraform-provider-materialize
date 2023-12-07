# Roles can be imported using the role id:
terraform import materialize_role.example_role <region>:<role_id>

# Role id and information be found in the `mz_catalog.mz_roles` table
# The region is the region where the database is located (e.g. aws/us-east-1)
