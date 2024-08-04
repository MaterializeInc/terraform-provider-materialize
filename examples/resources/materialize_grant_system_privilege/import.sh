# Grants can be imported using the concatenation of
# GRANT SYSTEM, the id of the role and the privilege
terraform import materialize_grant_system_privilege.example <region>:GRANT SYSTEM|<role_id>|<privilege>

# The region is the region where the database is located (e.g. aws/us-east-1)
