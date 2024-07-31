# Retrieve the SSO Configuration ID using the materialize_sso_config data source
# Example of using the data source in your configuration:
#
# data "materialize_sso_config" "all" {}
#
# output "sso_config_output" {
#     value = data.materialize_sso_config.all
# }
#
# The SSO Configuration ID can be retrieved using the following command:
# terraform output sso_config_output

# Import command:
terraform import materialize_sso_domain.example <sso_config_id>:<domain.com>

# Note: Replace <sso_config_id> with the actual ID of the SSO configuration
# and <domain.com> with the actual domain you want to import
