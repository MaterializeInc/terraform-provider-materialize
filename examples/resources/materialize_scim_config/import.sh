# Retrieve the SCIM Configuration ID using the materialize_scim_configs data source
# Example of using the data source in your configuration:
#
# data "materialize_scim_configs" "all" {}
#
# output "scim_config_output" {
#     value = data.materialize_scim_configs.all
# }
#
# The ID can be retrieved using the following command:
# terraform output scim_config_output

# Import command:
terraform import materialize_scim_config.example_scim_config <scim_config_id>

# Note: Replace <scim_config_id> with the actual ID of your SCIM configuration
