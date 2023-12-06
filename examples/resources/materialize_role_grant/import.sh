#Grants can be imported using the concatenation of ROLE MEMBER, the id of the role and id of the member 
terraform import materialize_role_grant.example <region>:ROLE MEMBER|<role_id>|<member_id>

# The region is the region where the database is located (e.g. aws/us-east-1)
