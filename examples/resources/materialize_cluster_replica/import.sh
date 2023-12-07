# Cluster replicas can be imported using the cluster replica id:
terraform import materialize_cluster_replica.example_1_cluster_replica <region>:<cluster_replica_id>

# Cluster replica id and information be found in the `mz_catalog.mz_cluster_replicas` table
# The region is the region where the database is located (e.g. aws/us-east-1)
