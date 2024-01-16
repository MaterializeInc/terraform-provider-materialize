resource "materialize_sso_default_roles" "example_sso_default_roles" {
  sso_config_id = materialize_sso_configuration.example_sso_config.id
  roles         = ["Admin"]
}

resource "materialize_sso_default_roles" "example_sso_default_roles_2" {
  sso_config_id = materialize_sso_configuration.example_sso_config.id
  roles         = ["Admin", "Member"]
}
