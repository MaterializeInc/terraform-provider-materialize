resource "materialize_cluster" "cluster" {
  name    = "cluster"
  comment = "cluster comment"
}

resource "materialize_cluster" "cluster_source" {
  name = "cluster_sources"
}

resource "materialize_cluster" "cluster_sink" {
  name = "cluster_sinks"
  size = "3xsmall"
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

resource "materialize_cluster_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
}

resource "materialize_cluster" "managed_cluster" {
  name                          = "managed_cluster"
  replication_factor            = 2
  size                          = "3xsmall"
  introspection_interval        = "1s"
  introspection_debugging       = true
  idle_arrangement_merge_effort = 2
  disk                          = true
}

data "materialize_cluster" "all" {}

data "materialize_current_cluster" "default" {}
