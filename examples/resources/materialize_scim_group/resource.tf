# Create a SCIM group
resource "materialize_scim_group" "example_scim_group" {
  name        = "example-scim-group"
  description = "Example SCIM group"
}
