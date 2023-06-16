resource "materialize_role" "role_1" {
  name           = "role_1"
  create_role    = false
  create_db      = true
  create_cluster = false
}

resource "materialize_role" "role_2" {
  name           = "role_2"
  create_role    = true
  create_db      = false
  create_cluster = true
}

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
