resource "materialize_user" "example_user" {
  for_each = toset(["1", "2", "3", "4", "5"])
  email    = "example-user${each.key}@example.com"
  roles    = ["Member", "Admin"]
}
