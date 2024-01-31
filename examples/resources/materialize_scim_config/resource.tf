# Create a SCIM config
resource "materialize_scim_config" "example_scim_config" {
  connection_name = "example_connection"
  source          = "okta"
}
