---
page_title: "Provider: Materialize"
description: Manage Materialize with Terraform.
---

# Materialize provider

This repository contains a Terraform provider for managing resources in a [Materialize](https://materialize.com/) account.

## Example provider configuration

Configure the provider by adding the following block to your Terraform project:

```terraform
# Configuration-based authentication
provider "materialize" {
  host     = var.materialize_host     # optionally use MZ_HOST env var
  user     = var.materialize_user     # optionally use MZ_USER env var
  password = var.materialize_password # optionally use MZ_PASSWORD env var
  port     = var.materialize_port     # optionally use MZ_PORT env var
  database = var.materialize_database # optionally use MZ_DATABASE env var
}
```

## Schema

* `host` (String) Materialize host. Can also come from the `MZ_HOST` environment variable.
* `user` (String) Materialize user. Can also come from the `MZ_USER` environment variable.
* `password` (String, Sensitive) Materialize host. Can also come from the `MZ_PASSWORD` environment variable.
* `port` (Number) The Materialize port number to connect to at the server host. Can also come from the `MZ_PORT` environment variable. Defaults to 6875.
* `database` (String) The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.

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
