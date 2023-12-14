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
  password       = var.materialize_password # optionally use MZ_PASSWORD env var
  default_region = "aws/us-east-1"          # optionally use MZ_REGION env var
}
```

## Schema

* `password` (String, Sensitive) Materialize App Password. Can also come from the `MZ_PASSWORD` environment variable.
* `database` (String, Optional) The Materialize database. Can also come from the `MZ_DATABASE` environment variable. Defaults to `materialize`.
* `default_region` (String, Optional) The Materialize AWS region. Can also come from the `MZ_DEFAULT_REGION` environment variable. Defaults to `aws/us-east-1`.

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
