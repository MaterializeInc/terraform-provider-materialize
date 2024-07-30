resource "materialize_cluster" "cluster" {
  name    = "cluster"
  comment = "cluster comment"
}

# Create in separate region
resource "materialize_cluster" "cluster_us_west" {
  name    = "cluster"
  comment = "cluster comment"
  region  = "aws/us-west-2"
}

resource "materialize_cluster" "cluster_source" {
  name = "cluster_sources"
}

# Create in separate region
resource "materialize_cluster" "cluster_source_us_west" {
  name   = "cluster_sources"
  region = "aws/us-west-2"
}

resource "materialize_cluster" "cluster_sink" {
  name = "cluster_sinks"
  size = "3xsmall"
}

# Create in separate region
resource "materialize_cluster" "cluster_sink_us_west" {
  name   = "cluster_sinks"
  size   = "3xsmall"
  region = "aws/us-west-2"
}

resource "materialize_cluster" "cluster_by_name" {
  name             = "cluster_by_name"
  size             = "25cc"
  identify_by_name = true
}

resource "materialize_cluster" "scheduling_cluster" {
  name = "scheduling_cluster"
  size = "25cc"
  scheduling {
    on_refresh {
      enabled                 = true
      hydration_time_estimate = "1 hour"
    }
  }
}

# Create in separate region
resource "materialize_cluster" "scheduling_cluster_us_west" {
  name   = "scheduling_cluster"
  size   = "25cc"
  region = "aws/us-west-2"
  scheduling {
    on_refresh {
      enabled                 = true
      hydration_time_estimate = "1 hour"
    }
  }
}

resource "materialize_cluster" "no_replication" {
  name               = "no_replication"
  size               = "25cc"
  replication_factor = 0
}

# Create in separate region
resource "materialize_cluster" "no_replication_us_west" {
  name               = "no_replication"
  size               = "25cc"
  replication_factor = 0
  region             = "aws/us-west-2"
}

resource "materialize_cluster_grant" "cluster_grant_usage" {
  role_name    = materialize_role.role_1.name
  privilege    = "USAGE"
  cluster_name = materialize_cluster.cluster.name
}

# Create in separate region
resource "materialize_cluster_grant" "cluster_grant_usage_us_west" {
  role_name    = materialize_role.role_1_us_west.name
  privilege    = "USAGE"
  cluster_name = materialize_cluster.cluster_us_west.name
  region       = "aws/us-west-2"
}

resource "materialize_cluster_grant" "cluster_grant_create" {
  role_name    = materialize_role.role_2.name
  privilege    = "CREATE"
  cluster_name = materialize_cluster.cluster_source.name
}

# Create in separate region
resource "materialize_cluster_grant" "cluster_grant_create_us_west" {
  role_name    = materialize_role.role_2_us_west.name
  privilege    = "CREATE"
  cluster_name = materialize_cluster.cluster_source_us_west.name
  region       = "aws/us-west-2"
}

resource "materialize_cluster_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
}

# Create in separate region
resource "materialize_cluster_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee_us_west.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target_us_west.name
  region           = "aws/us-west-2"
}

resource "materialize_cluster" "managed_cluster" {
  name                    = "managed_cluster"
  replication_factor      = 2
  size                    = "25cc"
  introspection_interval  = "1s"
  introspection_debugging = true
  disk                    = true
}

# Create in separate region
resource "materialize_cluster" "managed_cluster_us_west" {
  name                    = "managed_cluster"
  replication_factor      = 2
  size                    = "25cc"
  introspection_interval  = "1s"
  introspection_debugging = true
  disk                    = true
  region                  = "aws/us-west-2"
}

data "materialize_cluster" "all" {}

data "materialize_current_cluster" "quickstart" {}
