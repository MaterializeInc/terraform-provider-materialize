# Create a service user and app password
resource "materialize_role" "production_dashboard" {
  name = "svc_production_dashboard"
}
resource "materialize_app_password" "production_dashboard_app_password" {
  name  = "production_dashboard_app_password"
  type  = "service"
  user  = materialize_role.production_dashboard.name
  roles = ["Member"]
}
resource "materialize_database_grant" "database_grant_usage" {
  role_name     = materialize_role.production_dashboard.name
  privilege     = "USAGE"
  database_name = "production_analytics"
}

# Create a personal app password for the current user
resource "materialize_app_password" "example_app_password" {
  name = "example_app_password_name"
}
