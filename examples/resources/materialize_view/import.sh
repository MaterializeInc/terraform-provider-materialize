# Views can be imported using the view id:
terraform import materialize_view.example_view <region>:<view_id>

# View id and information be found in the `mz_catalog.mz_views`
# The region is the region where the database is located (e.g. aws/us-east-1)
