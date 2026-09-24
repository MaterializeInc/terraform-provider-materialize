# Copies Organization Member permissions when created. When assigned to a user,
# this role maps to the database role below through a JWT claim. Grant database
# privileges separately.
resource "materialize_organization_role" "reader" {
  name           = "analytics_reader"
  base_role_name = "Member"
}

resource "materialize_role" "reader" {
  name = materialize_organization_role.reader.key
}
