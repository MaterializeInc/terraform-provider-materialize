# Contributing

## Developing the provider

### Requirements

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (see
[Requirements](#requirements) above).

### Building The Provider

Clone the repository:

```
git clone https://github.com/MaterializeInc/terraform-provider-materialize.git
cd terraform-provider-materialize
```

Compile the provider

```bash
make install
```

### Generating the documentation

The documentation is generated from the provider's schema. To generate the documentation, run:

```bash
terraform fmt -recursive ./examples/
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
```

## Testing the Provider

### Running the unit tests

To run the full suite of acceptance tests, run:

```bash
make testacc
```

### Running the integration tests

To run the full suite of integration tests run:

```bash
# Start all containers
docker-compose up -d

# Run the tests
docker exec provider terraform init
docker exec provider terraform apply -auto-approve -compact-warnings
docker exec provider terraform plan -detailed-exitcode
docker exec provider terraform destroy -auto-approve -compact-warnings

# Stop all containers
docker-compose down -v
```

> Note: You might have to delete the `integration/.terraform`, `integration/.terraform.lock.hcl` and `integration/terraform.tfstate*` files before running the tests.

### Debugging
Terraform has detailed logs that you can enable by setting the `TF_LOG` environment variable to any value. Enabling this setting causes detailed logs to appear on `stderr`.

## Adding dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using
Go modules.

To add a new dependency:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Cutting a release

```bash
git tag -a vX.Y.Z -m vX.Y.Z
git push origin vX.Y.Z
```

[Materialize]: https://materialize.com
