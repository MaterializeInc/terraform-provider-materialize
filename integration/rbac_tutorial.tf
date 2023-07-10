# Step 2. Create a new role
resource "materialize_role" "dev_role" {
  name = "dev_role"
}

resource "materialize_role" "user" {
  name = "user"
}

# Step 4. Grant privileges to the role
resource "materialize_table_grant" "dev_role_table_grant_select" {
  role_name     = materialize_role.dev_role.name
  privilege     = "SELECT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_table_grant" "dev_role_table_grant_insert" {
  role_name     = materialize_role.dev_role.name
  privilege     = "INSERT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_table_grant" "dev_role_table_grant_update" {
  role_name     = materialize_role.dev_role.name
  privilege     = "UPDATE"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_schema_grant" "dev_role_schema_grant_usage" {
  role_name     = materialize_role.dev_role.name
  privilege     = "USAGE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
}

# TODO: Implement privilege ALL
resource "materialize_database_grant" "dev_role_database_grant_usage" {
  role_name     = materialize_role.dev_role.name
  privilege     = "USAGE"
  database_name = materialize_database.database.name
}

resource "materialize_database_grant" "dev_role_database_grant_create" {
  role_name     = materialize_role.dev_role.name
  privilege     = "CREATE"
  database_name = materialize_database.database.name
}

resource "materialize_cluster_grant" "dev_role_cluster_grant_usage" {
  role_name    = materialize_role.dev_role.name
  privilege    = "USAGE"
  cluster_name = materialize_cluster.cluster.name
}

resource "materialize_cluster_grant" "dev_role_cluster_grant_create" {
  role_name    = materialize_role.dev_role.name
  privilege    = "CREATE"
  cluster_name = materialize_cluster.cluster_source.name
}

# Step 5. Assign the role to a user
resource "materialize_grant_role" "dev_role_grant_user" {
  role_name   = materialize_role.dev_role.name
  member_name = materialize_role.user.name
}

# Step 6. Create a second role
resource "materialize_role" "qa_role" {
  name = "qa_role"
}

resource "materialize_grant_system_privilege" "qa_role_system_createdb" {
  role_name = materialize_role.dev_role.name
  privilege = "CREATEDB"
}

resource "materialize_database_grant" "qa_role_database_grant_usage" {
  role_name     = materialize_role.qa_role.name
  privilege     = "USAGE"
  database_name = materialize_database.database.name
}

resource "materialize_database_grant" "qa_role_database_grant_create" {
  role_name     = materialize_role.qa_role.name
  privilege     = "CREATE"
  database_name = materialize_database.database.name
}

# Step 7. Add inherited privileges
resource "materialize_grant_role" "qa_role_grant_dev_role" {
  role_name   = materialize_role.qa_role.name
  member_name = materialize_role.dev_role.name
}