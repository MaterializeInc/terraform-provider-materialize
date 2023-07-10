resource "materialize_cluster" "cluster" {
  name = "cluster"
}

resource "materialize_cluster" "cluster_source" {
  name = "cluster_sources"
}

resource "materialize_cluster_grant" "cluster_grant_usage" {
  role_name    = materialize_role.role_1.name
  privilege    = "USAGE"
  cluster_name = materialize_cluster.cluster.name
}

resource "materialize_cluster_grant" "cluster_grant_create" {
  role_name    = materialize_role.role_2.name
  privilege    = "CREATE"
  cluster_name = materialize_cluster.cluster_source.name
}

data "materialize_cluster" "all" {}

data "materialize_current_cluster" "default" {}
