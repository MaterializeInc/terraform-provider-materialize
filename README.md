# Terraform Provider: Materialize

[![Slack Badge](https://img.shields.io/badge/Join%20us%20on%20Slack!-blueviolet?style=flat&logo=slack&link=https://materialize.com/s/chat)](https://materialize.com/s/chat)

This repository contains a Terraform provider for managing resources in a [Materialize](https://materialize.com/) account.

## Requirements

* Materialize >= 0.27
* [Terraform](https://www.terraform.io/downloads.html) >= 1.0.3
* (Development) [Go](https://golang.org/doc/install) >= 1.23

## Installation

The `materialize` provider is published to the [Terraform Registry](https://registry.terraform.io/providers/MaterializeInc/materialize/latest). To use it, add the following configuration to your Terraform settings:

```hcl
terraform {
  required_providers {
    materialize = {
      source = "MaterializeInc/materialize"
    }
  }
}
```

Configure the provider by adding the following block to your Terraform project:

```hcl
provider "materialize" {
  password       = "materialize_password"
  default_region = "aws/us-east-1"
  database       = "materialize"
}
```

Once you have configured the provider, you can start defining resources using Terraform. You can find examples on how to define resources in the [`examples`](./examples/) directory.

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
  sasl_username {
    text = "user"
  }
  sasl_password {
    name          = "kafka_password"
    database_name = "materialize"
    schema_name   = "public"
  }
  sasl_mechanisms = "SCRAM-SHA-256"
  progress_topic  = "example"
}
```

### Data sources

You can use data sources to retrieve information about existing resources. For example, to retrieve information about the existing sinks in your Materialize region, add the following data source definition to your Terraform project:

```hcl
# main.tf
data "materialize_connection" "all" {}

output name {
  value       = data.materialize_connection.all
}
```

### Importing existing resources

You can import existing resources into your Terraform state using the `terraform import` command. For this, you will need the `id` of the resource from the respective [`mz_catalog`](https://materialize.com/docs/sql/system-catalog/mz_catalog/) system table.

For example, to import an existing connection named `kafka_connection`, first add the resource definition to your Terraform project:

```hcl
# main.tf
resource "materialize_connection_kafka" "kafka_connection" {
  name = "kafka_connection"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
}
```

Then, look up the connection id (`connection_id`) in [`mz_connections`](https://materialize.com/docs/sql/system-catalog/mz_catalog/#mz_connections) and run:

```bash
terraform import materialize_connection_kafka.kafka_connection <connection_id>
```

After the import, you can check the state of the resource by running:

```bash
terraform state show materialize_connection_kafka.kafka_connection
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute to this provider.

## License

This provider is distributed under the [Apache License, Version 2.0](LICENSE).

[Materialize Cloud]: https://cloud.materialize.com
