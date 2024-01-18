data "materialize_sso_configuration" "all" {}

output "sso_configurations" {
  value = data.materialize_sso_configuration.all.sso_configurations
}
