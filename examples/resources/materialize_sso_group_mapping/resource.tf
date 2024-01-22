resource "materialize_sso_group_mapping" "example_sso_group_mapping" {
  group         = "admins"
  sso_config_id = materialize_sso_config.example_sso_config.id
  roles         = ["Admin"]
}
