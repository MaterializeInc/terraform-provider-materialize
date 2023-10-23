resource "materialize_cluster_replica" "cluster_replica_1" {
  name         = "r1"
  cluster_name = materialize_cluster.cluster.name
  size         = "2xsmall"
  comment      = "cluster replica comment"
}

resource "materialize_cluster_replica" "cluster_replica_2" {
  name                          = "r2"
  cluster_name                  = materialize_cluster.cluster.name
  size                          = "3xsmall"
  availability_zone             = "test2"
  introspection_interval        = "2s"
  introspection_debugging       = true
  idle_arrangement_merge_effort = 1
  disk                          = true
}

resource "materialize_cluster_replica" "cluster_replica_source" {
  name         = "r1"
  cluster_name = materialize_cluster.cluster_source.name
  size         = "3xsmall"
}

data "materialize_cluster_replica" "all" {}
