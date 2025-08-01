# =============================================================================
# Materialize Cloud (SaaS) Configuration
# =============================================================================
# Use this configuration for Materialize Cloud environments.
# This provides access to ALL provider resources including:
# - App passwords, users, SSO, SCIM resources
# - All database resources (clusters, sources, sinks, etc.)
#
provider "materialize" {
  password       = var.materialize_password # optionally use MZ_PASSWORD env var
  default_region = "aws/us-east-1"          # optionally use MZ_DEFAULT_REGION env var
}

# =============================================================================
# Self-Hosted Materialize Configuration
# =============================================================================
# Use this configuration for self-hosted Materialize instances.
# 
# ⚠️  IMPORTANT LIMITATIONS:
# - Frontegg-dependent resources are NOT available (app passwords, users, SSO, SCIM)
# - Only database resources are available (clusters, sources, sinks, schemas, etc.)
# - No organization or identity management features
#
provider "materialize" {
  host     = "materialized" # optionally use MZ_HOST env var
  port     = 6875           # optionally use MZ_PORT env var
  username = "materialize"  # optionally use MZ_USER env var
  database = "materialize"  # optionally use MZ_DATABASE env var
  password = ""             # optionally use MZ_PASSWORD env var
  sslmode  = "disable"      # optionally use MZ_SSLMODE env var
}

# =============================================================================
# Migration Note
# =============================================================================
# Switching between SaaS and self-hosted modes requires careful state file
# management as resource references and regional configurations differ between modes.
