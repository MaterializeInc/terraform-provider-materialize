resource "materialize_sso_domain" "example_sso_domain" {
  domain        = "example.com"
  sso_config_id = materialize_sso_config.example_sso_config.id
}
