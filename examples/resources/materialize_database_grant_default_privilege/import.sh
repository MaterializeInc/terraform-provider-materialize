# Grants can be imported using the concatenation of GRANT DEFAULT and the grantee id of the role.
# Optionally, you can include the target id and database id. The privilege is required.
terraform import materialize_database_grant_default_privilege.example <region>:GRANT DEFAULT|CONNECTION|<grantee_id>|<target_role_id>|<database_id>||<privilege>

# The region is the region where the database is located (e.g. aws/us-east-1)
