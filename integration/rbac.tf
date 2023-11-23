resource "materialize_database" "db1" {
  name = "db1"
}

resource "materialize_database" "db2" {
  name = "db2"
}

resource "materialize_schema" "schema1" {
  name          = "schema1"
  database_name = materialize_database.db1.name
}

resource "materialize_schema" "schema2" {
  name          = "schema2"
  database_name = materialize_database.db1.name
}

resource "materialize_schema" "schema3" {
  name          = "schema3"
  database_name = materialize_database.db2.name
}

resource "materialize_schema" "schema4" {
  name          = "schema4"
  database_name = materialize_database.db2.name
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

resource "materialize_role" "target_1" {
  name = "target_1"
}

resource "materialize_role" "target_2" {
  name = "target_2"
}

variable "database_grants" {
  type = map(any)
  default = {
    bi_usage_t1 = {
      grantee : "bi",
      privilege : "USAGE",
      target_role : "target_1",
    },
    bi_create_t1 = {
      grantee : "bi",
      privilege : "CREATE",
      target_role : "target_1",
    },
    bi_create_t2 = {
      grantee : "bi",
      privilege : "CREATE",
      target_role : "target_2",
    },
    de_usage_t2 = {
      grantee : "de",
      privilege : "USAGE",
      target_role : "target_2",
    },
  }
}

resource "materialize_database_grant_default_privilege" "complex" {
  for_each         = var.database_grants
  privilege        = each.value.privilege
  grantee_name     = each.value.grantee
  target_role_name = each.value.target_role

  depends_on = [
    materialize_role.bi,
    materialize_role.ds,
    materialize_role.de,
    materialize_role.target_1,
    materialize_role.target_2,
    materialize_database.db1,
    materialize_database.db2,
  ]
}

variable "schema_grants" {
  type = map(any)
  default = {
    bi_usage_db1 = {
      grantee : "bi",
      privilege : "USAGE",
      database : "db1",
    },
    bi_usage_db2 = {
      grantee : "bi",
      privilege : "USAGE",
      database : "db2",
    },
    de_usage_db1 = {
      grantee : "de",
      privilege : "USAGE",
      database : "db2",
    },
    bi_usage_nodb = {
      grantee : "bi",
      privilege : "USAGE",
      database : "",
    },
  }
}

resource "materialize_schema_grant_default_privilege" "complex" {
  for_each         = var.schema_grants
  privilege        = each.value.privilege
  grantee_name     = each.value.grantee
  target_role_name = materialize_role.target.name
  database_name    = each.value.database

  depends_on = [
    materialize_role.bi,
    materialize_role.ds,
    materialize_role.de,
    materialize_database.db1,
    materialize_database.db2,
    materialize_schema.schema1,
    materialize_schema.schema2,
    materialize_schema.schema3,
    materialize_schema.schema4,
  ]
}

variable "table_grants" {
  type = map(any)
  default = {
    bi_select_db1_s1 = {
      grantee : "bi",
      privilege : "SELECT",
      database : "db1",
      schema : "schema1"
    },
    bi_insert_db1_s2 = {
      grantee : "bi",
      privilege : "INSERT",
      database : "db1",
      schema : "schema1"
    },
    ds_select_db2_s3 = {
      grantee : "ds",
      privilege : "SELECT",
      database : "db1",
      schema : "schema2"
    },
    de_select_db1_s3 = {
      grantee : "de",
      privilege : "SELECT",
      database : "db2",
      schema : "schema3"
    },
    de_update_db1_s1 = {
      grantee : "de",
      privilege : "UPDATE",
      database : "db1",
      schema : "schema1",
    },
    de_update_db1_nos = {
      grantee : "de",
      privilege : "UPDATE",
      database : "db1",
      schema : "",
    },
    de_update_nodb_nos = {
      grantee : "de",
      privilege : "UPDATE",
      database : "",
      schema : "",
    },
    de_update_db1_public = {
      grantee : "de",
      privilege : "UPDATE",
      database : "db1",
      schema : "public",
    },
  }
}

resource "materialize_table_grant_default_privilege" "complex" {
  for_each         = var.table_grants
  privilege        = each.value.privilege
  grantee_name     = each.value.grantee
  target_role_name = materialize_role.target.name
  database_name    = each.value.database
  schema_name      = each.value.schema

  depends_on = [
    materialize_role.bi,
    materialize_role.ds,
    materialize_role.de,
    materialize_database.db1,
    materialize_database.db2,
    materialize_schema.schema1,
    materialize_schema.schema2,
    materialize_schema.schema3,
    materialize_schema.schema4,
  ]
}

resource "materialize_role" "op" {
  name = "op"
}

# Non-deterministic grants
resource "materialize_table_grant_default_privilege" "base_insert" {
  privilege        = "INSERT"
  grantee_name     = materialize_role.de.name
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name
}

resource "materialize_table_grant_default_privilege" "base_update" {
  privilege        = "UPDATE"
  grantee_name     = materialize_role.de.name
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name
}

resource "materialize_table_grant_default_privilege" "duplicate_udpate" {
  privilege        = "UPDATE"
  grantee_name     = materialize_role.de.name
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  schema_name      = materialize_schema.schema.name
}
