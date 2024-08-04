# Grants can be imported using the concatenation of
# GRANT, the object type, the id of the object, the id of the role and the privilege
terraform import materialize_cluster_grant.example <region>:GRANT|CLUSTER|<cluster_id>|<role_id>|<privilege>

# The region is the region where the database is located (e.g. aws/us-east-1)
