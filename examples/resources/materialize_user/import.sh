# Retrieve the User ID using the materialize_user data source
# Example of using the data source in your configuration:
#
# data "materialize_user" "example_user" {
#   email = "example@example.com"
# }
#
# output "user_output" {
#   value = data.materialize_user.example_user
# }
#
# The User ID can be retrieved using the following command:
# terraform output -json user_output | jq '.id'

# Import command:
terraform import materialize_user.example_user <user_id>

# Note: Replace <user_id> with the actual ID of the user you want to import
# You can find the user ID by querying the data source with the user's email
