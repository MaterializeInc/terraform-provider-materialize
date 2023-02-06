# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

## Test sample configuration

First, build and install the provider (in the root directory).

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Create a file `locals.tf` with your Materialize connection details
*locals.tf*
```terraform
locals {
  host     = "xxx.us-east-1.aws.materialize.cloud"
  username = "{YOUR Username}"
  password = "{YOUR App Password}"
  port     = 6875
  database = "materialize"
}
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
