resource "materialize_user" "example_user" {
  email = "example-user@example.com"
  roles = ["Member", "Admin"]
}
