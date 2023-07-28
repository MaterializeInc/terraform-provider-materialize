#Grants can be imported using the concatenation of GRANT DEFAULT, the grantee id of the role
#Optionally you can include the target id, database id. The privilege is required 
terraform import materialize_database_grant_default_privilege.example GRANT DEFAULT|CONNECTION|<grantee_id>|<target_role_id>|<database_id>||<privilege>
