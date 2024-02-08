# Contributing

## Developing the provider

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

### Generating documentation

The documentation is generated from the provider's schema. To generate the documentation, run:

```bash
make docs
```

## Testing the Provider

### Running unit tests

To run the unit tests run:

```bash
make test
```

### Running accpetance tests

To run the acceptance tests which will simulate running Terraform commands you will need to set the necessary envrionment variables and start the docker compose:

```bash
# Start all containers
docker-compose up -d --build
```

Add the following to your `hosts` file so that the provider can connect to the mock services:

```
127.0.0.1 materialized frontegg cloud
```

You can then run the acceptance tests:

```bash
make testacc
```

### Running integration tests

To run the full integration project, set the necessary env variables and start the docker compose similar to the acceptance tests. Then to interact with the provider you can run:

```bash
# Run the tests
docker exec provider terraform init
docker exec provider terraform apply -auto-approve -compact-warnings
docker exec provider terraform plan -detailed-exitcode
docker exec provider terraform destroy -auto-approve -compact-warnings
```

> Note: You might have to delete the `integration/.terraform`, `integration/.terraform.lock.hcl` and `integration/terraform.tfstate*` files before running the tests.

### Debugging
Terraform has detailed logs that you can enable by setting the `TF_LOG` environment variable to any value. Enabling this setting causes detailed logs to appear on `stderr`.

## Adding a Feature to the Provider

If you add a feature in Materialize, eventually it will need to be added to the Terraform provider. Here is a quick guide on how to update the provider.

Say we wanted to add `size` to the clusters.

### Step 1 - Update the query builder and query

In the [materialize package](https://github.com/MaterializeInc/terraform-provider-materialize/tree/main/pkg/materialize) find the corresponding resource. Within that file add the new feature to the builder:

```go
type ClusterBuilder struct {
	ddl                        Builder
	clusterName                string
	replicationFactor          int
	size                       string // Add new field
}
```

You can then update the `Create` method and, if necessary, add a method for handling any updates.

Next you can update the query that Terraform will run to find that feature:

```go
type ClusterParams struct {
	ClusterId         sql.NullString `db:"id"`
	ClusterName       sql.NullString `db:"name"`
	Managed           sql.NullBool   `db:"managed"`
	Size              sql.NullString `db:"size"`// Add new field
}

var clusterQuery = NewBaseQuery(`
	SELECT
		mz_clusters.id,
		mz_clusters.name,
		mz_clusters.managed,
		mz_clusters.size // Add new field
	FROM mz_clusters`)
```

After you update the query. You will also need to update the mock query in the [testhelpers package](https://github.com/MaterializeInc/terraform-provider-materialize/blob/main/pkg/testhelpers/mock_scans.go) so the tests will pass.

### Step 2 - Update the resource

In the [resources package](https://github.com/MaterializeInc/terraform-provider-materialize/tree/main/pkg/resources) find the corresponding resource. Within that file add the new attribute to the Terraform schema:

```go
var clusterSchema = map[string]*schema.Schema{
	"name": ObjectNameSchema("cluster", true, true),
	"size": {
		Description: "The size of the cluster.",
		Optional:    true,
	},
	"region": RegionSchema(),
}
```

You can then update the read context `clusterRead`:

```go
if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
}
```

And the create context `clusterCreate`:

```go
if v, ok := d.GetOk("size"); ok {
		b.Size(v.(string))
}
```

If the resource can be updated we would also have to change the update context `clusterUpdate`:

```go
if d.HasChange("size") {
		metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
		if err != nil {
			return diag.FromErr(err)
		}
		_, newSize := d.GetChange("size")
		b := materialize.NewClusterBuilder(metaDb, o)
		if err := b.Resize(newSize.(string)); err != nil {
				return diag.FromErr(err)
		}
}
```

### Step 3 - Update datasource

In the [datasources package](https://github.com/MaterializeInc/terraform-provider-materialize/tree/main/pkg/datasources) find the corresponding resource. Within that file add the new field the `Schema` for `Cluster`:

```go
Schema: map[string]*schema.Schema{
	"clusters": {
		Type:        schema.TypeList,
		Computed:    true,
		Description: "The clusters in the account",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"size": { // Add new field
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	},
	"region": {
		Type:     schema.TypeString,
		Computed: true,
	},
},
```

And finally update the mapping in `clusterRead`:

```go
for _, p := range dataSource {
		clusterMap := map[string]interface{}{}

		clusterMap["id"] = p.ClusterId.String
		clusterMap["name"] = p.ClusterName.String
		clusterMap["size"] = p.Size.String // Add new field

		clusterFormats = append(clusterFormats, clusterMap)
	}
```

## Cutting a release

To cut a new release of the provider, create a new tag and push that tag. This will trigger a GitHub Action to generate the artifacts necessary for the Terraform Registry.

```bash
git tag -a vX.Y.Z -m vX.Y.Z
git push origin vX.Y.Z
```

[Materialize]: https://materialize.com
