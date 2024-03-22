# ALTER ROLE some_role SET transaction_isolation = 'strict serializable';
resource "materialize_role" "example_role" {
  name = "some_role"
}

resource "materialize_role_parameter" "example_role_parameter" {
  role_name      = materialize_role.example_role.name
  variable_name  = "transaction_isolation"
  variable_value = "strict serializable"
}
