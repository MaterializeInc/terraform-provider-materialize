# Terraform Provider Materialize

> **Warning**
> The Terraform Provider for Materialize is under active development.

This repository contains a Terraform provider for the [Materialize platform](https://cloud.materialize.com/).

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) >= 1.0.3
* [Go](https://golang.org/doc/install) >= 1.16

## Installation

To use the provider, add the following configuration to your Terraform settings:

```hcl
terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}
```

Configure the provider by adding the following block to your Terraform project:

```hcl
provider "materialize" {
  host     = "materialized_hostname"
  username = "materialize_username"
  password = "materialize_password"
  port     = 6875
  database = "materialize"
}
```

Once you have configured the provider, you can start defining resources using Terraform. You can find examples of how to define resources in the [`examples`](./examples/) directory.

## Usage

### Managing resources

You can manage resources using the `terraform apply` command. For example, to create a new connection named `kafka_connection`, add the following resource definition to your Terraform project:

```hcl
# main.tf
resource "materialize_connection_kafka" "kafka_connection" {
  name = "kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
  sasl_username = "example"
  sasl_password {
    name          = "kafka_password"
    database_name = "materialize"
    schema_name   = "public"
  }
  sasl_mechanisms = "SCRAM-SHA-256"
  progress_topic  = "example"
}
```

Then, run apply the changes:

```bash
terraform apply
```

### Data sources

You can use data sources to retrieve information about existing resources. For example, to retrieve information about the existing sinks in your Materialize instance, add the following data source definition to your Terraform project:

```hcl
# main.tf
data "materialize_connection" "all" {}

output name {
  value       = data.materialize_connection.all
}
```

Then, check the Terraform plan:

```bash
terraform plan
```

### Importing existing resources

You can import existing resources into your Terraform state using the `terraform import` command. For example, to import an existing connection named `kafka_connection`, first add the resource definition to your Terraform project:

```hcl
# main.tf
resource "materialize_connection_kafka" "kafka_connection" {
  name = "kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
}
```

Then, run the following command:

```bash
terraform import materialize_connection_kafka.kafka_connection CONNECTION_ID
```

After the import, you can check the state of the resource by running the following command:

```bash
terraform state show materialize_connection_kafka.kafka_connection
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute to this provider.

## License

This provider is distributed under the [Apache License, Version 2.0](LICENSE).

[Materialize Cloud]: https://cloud.materialize.com
