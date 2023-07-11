#Grants can be imported using the concatenation of GRANT DEFAULT, the grantee id of the role
#Optionally you can include the target id. The privilege is required 
terraform import materialize_cluster_grant_default_privilege.example GRANT DEFAULT|CLUSTER|<grantee_id>|<target_role_id>|||<privilege>
