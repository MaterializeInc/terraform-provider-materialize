resource "materialize_role" "role_1" {
  name = "role-1"
}

resource "materialize_role" "role_2" {
  name = "role-2"
}

resource "materialize_role" "grantee" {
  name = "grantee"
}

resource "materialize_role" "target" {
  name = "target"

  session_variable {
    name  = "datestyle"
    value = "ISO, MDY"
  }
  session_variable {
    name  = "emit_timestamp_notice"
    value = "true"
  }
  session_variable {
    name  = "max_query_result_size"
    value = "1073741824"
  }
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

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
