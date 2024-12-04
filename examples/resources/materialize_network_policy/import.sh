# Network policies can be imported using the `terraform import` command.
terraform import materialize_network_policy.example_network_policy <region>:<network_policy_id>

# The network_policy_id is the ID of the network policy.
# You can find it from the `mz_internal.mz_network_policy_rules` table in the materialize
# The region is the region where the database is located (e.g. aws/us-east-1)
