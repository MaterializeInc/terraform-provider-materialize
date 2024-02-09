# ALTER SYSTEM SET cluster TO 'cluster_name';
resource "materialize_system_parameter" "example_system_parameter" {
  name  = "cluster"
  value = "cluster_name"
}
