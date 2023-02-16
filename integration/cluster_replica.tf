resource "materialize_cluster_replica" "cluster_replica_1" {
  name         = "r1"
  cluster_name = materialize_cluster.cluster.name
  size         = "8"
}

resource "materialize_cluster_replica" "cluster_replica_2" {
  name                          = "r2"
  cluster_name                  = materialize_cluster.cluster.name
  size                          = "4"
  availability_zone             = "us-east-1"
  introspection_interval        = "2s"
  introspection_debugging       = true
  idle_arrangement_merge_effort = 1
}

resource "materialize_cluster_replica" "cluster_replica_source" {
  name         = "r1"
  cluster_name = materialize_cluster.cluster_source.name
  size         = "8"
}