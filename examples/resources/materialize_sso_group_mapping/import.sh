# Retrieve the SSO Configuration ID and Group Mapping ID using the materialize_sso_config data source
# Example of using the data source in your configuration:
#
# data "materialize_sso_config" "all" {}
#
# output "sso_config_output" {
#     value = data.materialize_sso_config.all
# }
#
# The SSO Configuration ID and Group Mapping ID can be retrieved using the following command:
# terraform output sso_config_output

# Import command:
terraform import materialize_sso_group_mapping.example <sso_config_id>:<sso_group_mapping_id>

# Note: Replace <sso_config_id> with the actual ID of the SSO configuration
# and <sso_group_mapping_id> with the actual ID of the group mapping you want to import
