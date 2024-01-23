resource "materialize_sso_config" "example_sso_config" {
  enabled            = true
  sso_endpoint       = "https://sso.example2.com"
  public_certificate = "PUBLIC_CERTIFICATE"
  sign_request       = true
  type               = "saml"
  oidc_client_id     = "client-id"
  oidc_secret        = "client-secret"
}
