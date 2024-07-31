# Grants can be imported using the concatenation of GRANT DEFAULT, the grantee id of the role
# Optionally you can include the target id, database id and schema id. The privilege is required
terraform import materialize_secret_grant_default_privilege.example <region>:GRANT DEFAULT|SECRET|<grantee_id>|<target_role_id>|<database_id>|<schema_id>|<privilege>

# The region is the region where the database is located (e.g. aws/us-east-1)
