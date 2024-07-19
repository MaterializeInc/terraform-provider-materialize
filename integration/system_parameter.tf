resource "materialize_system_parameter" "default_cluster" {
  name  = "max_tables"
  value = "2000"
}

# Create in separate region
resource "materialize_system_parameter" "default_cluster_us_west" {
  name   = "max_tables"
  value  = "2000"
  region = "aws/us-west-2"
}

resource "materialize_system_parameter" "max_roles" {
  name  = "max_roles"
  value = "2000"
}

# Create in separate region
resource "materialize_system_parameter" "max_roles_us_west" {
  name   = "max_roles"
  value  = "2000"
  region = "aws/us-west-2"
}

data "materialize_system_parameter" "all" {}

data "materialize_system_parameter" "max_tables" {
  name = "max_tables"
}
