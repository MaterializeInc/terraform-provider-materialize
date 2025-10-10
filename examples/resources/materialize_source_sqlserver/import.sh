#!/bin/bash

# SQL Server sources can be imported using the source name
terraform import materialize_source_sqlserver.example <region>:<source_name>

# Example
terraform import materialize_source_sqlserver.example aws/us-east-1:my_sqlserver_source
