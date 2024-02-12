resource "materialize_system_parameter" "default_cluster" {
  name  = "max_tables"
  value = "2000"
}

resource "materialize_system_parameter" "max_roles" {
  name  = "max_roles"
  value = "2000"
}

data "materialize_system_parameter" "all" {}
