resource "materialize_role" "example_role" {
  name           = "example_role"
  create_role    = false
  create_db      = true
  create_cluster = false
}
