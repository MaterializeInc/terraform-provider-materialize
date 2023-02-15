resource "materialize_cluster_replica" "example_cluster_replica" {
  name         = "replica"
  cluster_name = "cluster"
  size         = "2xsmall"
}