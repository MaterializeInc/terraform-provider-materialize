resource "materialize_role" "role_1" {
  name = "role_1"
}

resource "materialize_role" "role_2" {
  name = "role_2"
}

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
