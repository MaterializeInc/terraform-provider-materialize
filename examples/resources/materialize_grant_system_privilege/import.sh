#Grants can be imported using the concatenation of GRANT SYSTEM, the id of the role and the privilege 
terraform import materialize_grant_system_privilege.example GRANT SYSTEM|<role_id>|<privilege>
