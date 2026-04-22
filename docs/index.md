---
page_title: "Provider: Materialize"
description: Manage Materialize with Terraform.
---

# Materialize provider

This repository contains a Terraform provider for managing resources in a [Materialize](https://materialize.com/) account.

## Example provider configuration

Configure the provider by adding the following block to your Terraform project:

```terraform
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
# Self-Hosted Materialize with OIDC/SSO Authentication
# =============================================================================
# Use this configuration when Self-Managed Materialize is deployed with
# `authenticatorKind: Oidc`. The `oidc_auth_enabled=true` connection option is
# REQUIRED — without it, Materialize falls back to password authentication.
#
# The `password` field should be an OIDC ID token obtained from your identity
# provider; the `username` should be the value of the JWT claim configured via
# `oidc_authentication_claim` (e.g. the user's email).
#
# See https://materialize.com/docs/security/self-managed/sso/ for details.
#
provider "materialize" {
  host     = "materialized"
  port     = 6875
  username = var.oidc_username # e.g. alice@your-org.com
  password = var.oidc_id_token # OIDC ID token from your IdP
  database = "materialize"
  options = {
    oidc_auth_enabled = "true"
  }
}

# =============================================================================
# Migration Note
# =============================================================================
# Switching between SaaS and self-hosted modes requires careful state file
# management as resource references and regional configurations differ between modes.
```

## ⚠️ Important: SaaS vs Self-Hosted Resources

**The provider supports two distinct configuration modes with different resource availability:**

### Materialize Cloud (SaaS) Mode
When configured with `password` and `default_region` (first example above), **all resources are available**.

### Self-Hosted Mode  
When configured with `host`, `username`, etc. (second example above), **some resources are NOT available**, including:

- `materialize_app_password`
- `materialize_user` 
- `materialize_sso_config` and related SSO resources
- `materialize_scim_config` and related SCIM resources

These organization and identity management resources depend on Frontegg (Materialize Cloud's identity provider) and will produce clear error messages if used in self-hosted mode.

**⚠️ Migration Warning:** Switching between SaaS and self-hosted modes requires careful state file management. We strongly recommend using consistent configuration mode from the beginning to avoid complex state migrations.

## Schema

* `password` (String, Sensitive) Materialize App Password (SaaS) or database password (self-hosted). Can also come from the `MZ_PASSWORD` environment variable.
* `default_region` (String, Optional) The Materialize AWS region (SaaS only). Can also come from the `MZ_DEFAULT_REGION` environment variable. Defaults to `aws/us-east-1`.
* `host` (String, Optional) The Materialize host (self-hosted only). Can also come from the `MZ_HOST` environment variable.
* `port` (Number, Optional) The Materialize port (self-hosted only). Can also come from the `MZ_PORT` environment variable. Defaults to `6875`.
* `username` (String, Optional) The database username (self-hosted only). Can also come from the `MZ_USER` environment variable. Defaults to `materialize`.
* `database` (String, Optional) The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.
* `sslmode` (String, Optional) SSL mode (self-hosted only). Can also come from the `MZ_SSLMODE` environment variable. Defaults to `require`.
* `options` (Map of String, Optional) Additional Postgres connection options forwarded in the `options` connection string parameter as `--key=value` flags. Useful for session-level settings such as `cluster`, `search_path`, or `oidc_auth_enabled` (required for OIDC/SSO authentication). The `transaction_isolation` and `application_name` keys are reserved and managed by the provider.

## Authenticating via OIDC/SSO (self-hosted)

When Self-Managed Materialize is configured for [OIDC authentication](https://materialize.com/docs/security/self-managed/sso/),
connect by passing `oidc_auth_enabled = "true"` via the `options` map and using
an OIDC ID token as the password:

```terraform
provider "materialize" {
  host     = "materialized"
  port     = 6875
  username = var.oidc_username # e.g. the value of the `oidc_authentication_claim`
  password = var.oidc_id_token # an OIDC ID token from your IdP
  database = "materialize"
  options = {
    oidc_auth_enabled = "true"
  }
}
```

**Token lifetime:** Materialize validates the OIDC token at connection time
only. If a single `terraform apply` outlives the token's expiry and the
provider needs to reconnect, authentication will fail. Use a token with a
lifetime comfortably longer than your longest apply, or plan to rerun with a
fresh token.

## Order precedence

The Materialize provider will use the following order of precedence when determining which credentials to use:
1. Provider configuration
2. Environment variables

## Modules

To help with your projects, you can use these Materialize maintained Terraform modules for common configurations:

* [MSK Privatelink](https://registry.terraform.io/modules/MaterializeInc/msk-privatelink/aws/latest)
* [Kafka Privatelink](https://registry.terraform.io/modules/MaterializeInc/kafka-privatelink/aws/latest)
* [EC2 SSH Bastion](https://registry.terraform.io/modules/MaterializeInc/ec2-ssh-bastion/aws/latest)
* [RDS Postgres](https://registry.terraform.io/modules/MaterializeInc/rds-postgres/aws/latest)

## Getting support

If you run into a snag or need support as you use the provider, join the Materialize [Slack community](https://materialize.com/s/chat) or [open an issue](https://github.com/MaterializeInc/terraform-provider-materialize/issues)!
