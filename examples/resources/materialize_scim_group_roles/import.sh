# Retrieve the SCIM Group ID using the materialize_scim_groups data source
# Example of using the data source in your configuration:
#
# data "materialize_scim_groups" "all" {}
#
# output "scim_group_output" {
#     value = data.materialize_scim_groups.all
# }
#
# The ID can be retrieved using the following command:
# terraform output scim_group_output

# Import command:
terraform import materialize_scim_group_roles.example_scim_group_roles <scim_group_id>

# Note: Replace <scim_group_id> with the actual ID of the SCIM group
# whose roles you want to import
