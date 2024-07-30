# Clusters can be imported using the cluster id or name
terraform import materialize_cluster.example_cluster <region>:id:<cluster_id>

# To import using the cluster name, you need to set the `identify_by_name` attribute to true
terraform import materialize_cluster.example_cluster <region>:name:<cluster_name>

# Cluster id and information be found in the `mz_catalog.mz_clusters` table
# The region is the region where the database is located (e.g. aws/us-east-1)
