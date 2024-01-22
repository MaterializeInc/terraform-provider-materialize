data "materialize_sso_config" "all" {}

output "sso_configs" {
  value = data.materialize_sso_config.all.sso_configs
}
