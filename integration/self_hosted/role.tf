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

resource "materialize_role" "self_hosted_admin" {
  name      = "self_hosted_admin"
  password  = "secure_password_123"
  superuser = true
  login     = true
}

resource "materialize_role" "self_hosted_user" {
  name      = "self_hosted_user"
  password  = "user_password_456"
  superuser = false
  login     = true
}

resource "materialize_role" "self_hosted_login_user" {
  name     = "self_hosted_login_user"
  password = "login_password_789"
  login    = false
}

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}

data "materialize_role" "role_prefix" {
  like_pattern = "role-%"
  depends_on = [
    materialize_role.role_1,
    materialize_role.role_2,
  ]
}

data "materialize_role" "self_hosted_roles" {
  like_pattern = "self_hosted_%"
  depends_on = [
    materialize_role.self_hosted_admin,
    materialize_role.self_hosted_user,
    materialize_role.self_hosted_login_user,
  ]
}
