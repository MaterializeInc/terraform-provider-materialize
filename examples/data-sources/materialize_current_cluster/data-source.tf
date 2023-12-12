data "materialize_current_cluster" "current" {}

output "cluster_name" {
  value = data.materialize_current_cluster.current.name
}
