# Grant role the privilege USAGE for objects
resource "materialize_cluster_grant_default_privilege" "example" {
  grantee_name     = "grantee"
  privilege        = "USAGE"
  target_role_name = "target_role"
}
