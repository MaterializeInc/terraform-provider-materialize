#!/bin/bash

# Start SQL Server in the background
/opt/mssql/bin/sqlservr &

# Wait for SQL Server to start up
echo "Waiting for SQL Server to start..."
echo "SA_PASSWORD is set: ${SA_PASSWORD:+yes}"

# Try to find sqlcmd - check different possible locations
SQLCMD=""
if [ -f "/opt/mssql-tools18/bin/sqlcmd" ]; then
    SQLCMD="/opt/mssql-tools18/bin/sqlcmd"
elif [ -f "/opt/mssql-tools/bin/sqlcmd" ]; then
    SQLCMD="/opt/mssql-tools/bin/sqlcmd"
else
    echo "Error: sqlcmd not found!"
    exit 1
fi

echo "Using sqlcmd at: $SQLCMD"

# Wait longer for SQL Server to initialize properly
for i in {1..120}
do
    $SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -Q "SELECT 1" -C > /dev/null 2>&1
    if [ $? -eq 0 ]
    then
        echo "SQL Server started successfully after $i attempts"
        break
    else
        echo "Attempt $i: Not ready yet..."
        sleep 2
    fi
done

# Double-check that we can connect before running bootstrap
echo "Verifying SQL Server connection..."
$SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -Q "SELECT @@VERSION" -C
if [ $? -ne 0 ]; then
    echo "Error: Could not connect to SQL Server for bootstrap"
    exit 1
fi

# Run the bootstrap script
echo "Running bootstrap script..."
$SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -i /docker-entrypoint-initdb.d/sqlserver_bootstrap.sql -C -t 30

if [ $? -eq 0 ]; then
    echo "Bootstrap script completed successfully"
else
    echo "Error: Bootstrap script failed"
    exit 1
fi

# Keep the container running
wait
