resource "materialize_network_policy" "office_policy" {
  name = "office_access_policy"

  rule {
    name      = "minnesota"
    action    = "allow"
    direction = "ingress"
    address   = "2.3.4.5/32"
  }

  rule {
    name      = "new_york"
    action    = "allow"
    direction = "ingress"
    address   = "1.2.3.4/28"
  }

  comment = "Network policy for office locations"
}

# An initial `default` network policy will be created.
# This policy allows open access to the environment and can be altered by a `superuser`.
# Use the `ALTER SYSTEM SET network_policy TO 'office_access_policy'` command.
# Or the `materialize_system_parameter` resource to set the default network policy.
resource "materialize_system_parameter" "system_parameter" {
  name  = "network_policy"
  value = "office_access_policy"
}
