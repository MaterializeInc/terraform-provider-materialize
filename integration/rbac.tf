variable "teams" {
  type = map(any)
  default = {
    bi_select = {
      grantee : "bi",
      privilege : "SELECT",
    },
    "bi_insert" = {
      grantee : "bi",
      privilege : "INSERT",
    },
    "ds_select" = {
      grantee : "ds",
      privilege : "SELECT",
    },
    "de_select" = {
      grantee : "de",
      privilege : "SELECT",
    },
    "de_update" = {
      grantee : "de",
      privilege : "UPDATE",
    },
  }
}

resource "materialize_role" "bi" {
  name = "bi"
}

resource "materialize_role" "ds" {
  name = "ds"
}

resource "materialize_role" "de" {
  name = "de"
}

resource "materialize_table_grant_default_privilege" "complex" {
  for_each         = var.teams
  privilege        = each.value.privilege
  grantee_name     = each.value.grantee
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name

  depends_on = [materialize_role.bi, materialize_role.ds, materialize_role.de]
}

resource "materialize_role" "op" {
  name = "op"
}

# Non-deterministic grants
resource "materialize_table_grant_default_privilege" "base" {
  privilege        = "UPDATE"
  grantee_name     = materialize_role.de.name
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name
}

resource "materialize_table_grant_default_privilege" "duplicate" {
  privilege        = "UPDATE"
  grantee_name     = materialize_role.de.name
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name
}
