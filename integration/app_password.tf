resource "materialize_app_password" "example_password" {
  for_each = toset(["1", "2", "3", "4", "5"])
  name     = "example_password_${each.key}"
}
