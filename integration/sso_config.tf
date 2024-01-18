
resource "materialize_sso_configuration" "example_sso_config" {
  enabled           = true
  sso_endpoint       = "https://sso.example.com"
  public_certificate = "DUMMY_CERTIFICATE"
  sign_request       = true
  type              = "saml"
  oidc_client_id      = "client-id"
  oidc_secret        = "client-secret"
}

resource "materialize_sso_domain" "example_sso_domain" {
  domain = "bobbyiliev.com"
  sso_config_id = materialize_sso_configuration.example_sso_config.id
}

resource "materialize_sso_group_mapping" "example_sso_group_mapping" {
  group         = "admins"
  sso_config_id = materialize_sso_configuration.example_sso_config.id
  roles         = ["Admin"]
}

resource "materialize_sso_group_mapping" "example_sso_group_mapping_2" {
  group         = "members"
  sso_config_id = materialize_sso_configuration.example_sso_config.id
  roles         = ["Member"]
}

resource "materialize_sso_default_roles" "example_sso_default_roles" {
  sso_config_id = materialize_sso_configuration.example_sso_config.id
  roles         = ["Admin", "Member"]
}

data "materialize_sso_configuration" "all" {}
