resource "materialize_role" "role_1" {
  name    = "role-1"
  comment = "role 1 comment"
}

resource "materialize_role" "role_2" {
  name    = "role-2"
  comment = "role 2 comment"
}

resource "materialize_role" "grantee" {
  name    = "grantee"
  comment = "role grantee comment"
}

resource "materialize_role" "target" {
  name = "target"
}

resource "materialize_grant_system_privilege" "role_1_system_createrole" {
  role_name = materialize_role.role_1.name
  privilege = "CREATEROLE"
}

resource "materialize_grant_system_privilege" "role_1_system_createdb" {
  role_name = materialize_role.role_1.name
  privilege = "CREATEDB"
}

resource "materialize_grant_system_privilege" "role_1_system_createcluster" {
  role_name = materialize_role.role_1.name
  privilege = "CREATECLUSTER"
}

resource "materialize_cluster_grant" "cluster_grant_public" {
  role_name    = "PUBLIC"
  privilege    = "USAGE"
  cluster_name = materialize_cluster.cluster.name
}

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
