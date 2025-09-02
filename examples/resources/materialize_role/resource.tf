resource "materialize_role" "example_role" {
  name = "example_role"
}

resource "materialize_role" "admin_user" {
  name      = "admin_user"
  password  = var.admin_password
  superuser = true
}
