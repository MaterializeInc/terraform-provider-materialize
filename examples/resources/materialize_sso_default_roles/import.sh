# Retrieve the SSO Configuration ID using the materialize_sso_config data source
# Example of using the data source in your configuration:
#
# data "materialize_sso_config" "all" {}
#
# output "sso_config_output" {
#     value = data.materialize_sso_config.all
# }
#
# The SSO configuration ID can be retrieved using the following command:
# terraform output sso_config_output

# Import command:
terraform import materialize_sso_default_roles.example_role <configuration_id>

# Note: Replace <configuration_id> with the actual ID of the SSO configuration
# whose default roles you want to import
