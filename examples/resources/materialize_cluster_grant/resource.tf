# Grant USAGE to role example_role to cluster example_cluster
resource "materialize_cluster_grant" "cluster_grant_usage" {
  role_name    = "example_role"
  privilege    = "USAGE"
  cluster_name = "example_cluster"
}
