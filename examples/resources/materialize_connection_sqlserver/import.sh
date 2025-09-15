#!/bin/bash

# SQL Server connections can be imported using the connection name
terraform import materialize_connection_sqlserver.example <region>:<connection_name>

# Example
terraform import materialize_connection_sqlserver.example aws/us-east-1:my_sqlserver_connection 
