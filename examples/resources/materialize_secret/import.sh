# Secrets can be imported using the secret id:
terraform import materialize_secret.example_secret <region>:<secret_id>

# Secret id and information be found in the `mz_catalog.mz_secrets` table
# The region is the region where the database is located (e.g. aws/us-east-1)
