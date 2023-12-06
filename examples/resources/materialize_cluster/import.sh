# Clusters can be imported using the cluster id:
terraform import materialize_cluster.example_cluster <region>:<cluster_id>

# Cluster id and information be found in the `mz_catalog.mz_clusters` table
# The region is the region where the database is located (e.g. aws/us-east-1)
