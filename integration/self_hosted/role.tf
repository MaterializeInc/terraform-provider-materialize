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

# TODO: Configure the materialized image to enable authentication for self-hosted testing
# resource "materialize_role" "self_hosted_admin" {
#   name      = "self_hosted_admin"
#   password  = "secure_password_123"
#   superuser = true
# }

# resource "materialize_role" "self_hosted_user" {
#   name      = "self_hosted_user"
#   password  = "user_password_456"
#   superuser = false
# }

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
