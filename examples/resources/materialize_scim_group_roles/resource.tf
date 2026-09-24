# Configure SCIM provisioning first, then wait for the identity provider
# to push the group before applying this mapping.
data "materialize_scim_groups" "all" {}

resource "materialize_organization_role" "reader" {
  name           = "analytics_reader"
  base_role_name = "Member"
}

locals {
  analytics_groups = [
    for group in data.materialize_scim_groups.all.groups : group
    if group.name == "analytics-team" && contains(["scim", "scim2"], group.managed_by)
  ]
}

resource "materialize_scim_group_roles" "reader" {
  group_id = try(one(local.analytics_groups).id, "")
  # Member assigns the reserved Organization Member [MaterializePlatform] role.
  roles = ["Member", materialize_organization_role.reader.name]

  lifecycle {
    precondition {
      condition     = length(local.analytics_groups) == 1
      error_message = "Wait for exactly one SCIM group named analytics-team to be provisioned, then rerun Terraform."
    }
  }
}
