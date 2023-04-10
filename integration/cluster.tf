resource "materialize_cluster" "cluster" {
  name = "cluster"
}

resource "materialize_cluster" "cluster_source" {
  name = "cluster_sources"
}

data "materialize_cluster" "all" {}
