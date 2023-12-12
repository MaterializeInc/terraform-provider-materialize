data "materialize_current_database" "current" {}

output "database_name" {
  value = data.materialize_current_database.current.name
}
