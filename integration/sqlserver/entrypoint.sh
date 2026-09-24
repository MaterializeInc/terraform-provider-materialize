#!/bin/bash

# sqlservr occasionally core dumps a couple of seconds into startup on the CI
# runners. Backgrounding it meant nothing noticed: the loop below kept polling a
# dead process until the healthcheck budget ran out and the job failed with
# "container sqlserver is unhealthy". Keep the pid so the loop can tell a slow
# start from a dead server, and start it again if it died.
start_sqlservr() {
    /opt/mssql/bin/sqlservr &
    SQLSERVR_PID=$!
}

sqlservr_alive() {
    kill -0 "$SQLSERVR_PID" 2>/dev/null
}

restart_sqlservr_if_dead() {
    if sqlservr_alive; then
        return 0
    fi
    echo "sqlservr (pid $SQLSERVR_PID) exited before accepting connections; last errorlog lines:"
    tail -n 20 /var/opt/mssql/log/errorlog 2>/dev/null || echo "(no errorlog yet)"
    echo "Starting sqlservr again..."
    start_sqlservr
}

start_sqlservr

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

# Wait longer for SQL Server to initialize properly. A short login timeout keeps
# each attempt from eating most of the budget while the server is still coming up.
for i in {1..120}
do
    restart_sqlservr_if_dead
    $SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -Q "SELECT 1" -C -l 3 > /dev/null 2>&1
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

# Enabling CDC creates a capture job, which needs SQL Server Agent. The Agent
# only finishes starting after SQL Server begins accepting connections, and
# sp_cdc_enable_table fails with 14258 while it is still coming up. The
# bootstrap runs without -b so that failure is not fatal, which left CDC
# silently disabled and the healthcheck waiting forever.
echo "Waiting for SQL Server Agent..."
for i in {1..60}
do
    $SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -C -b -Q \
        "SET NOCOUNT ON; IF NOT EXISTS (SELECT 1 FROM sys.dm_server_services WHERE servicename LIKE 'SQL Server Agent%' AND status_desc = 'Running') RAISERROR('agent not ready', 16, 1);" > /dev/null 2>&1
    if [ $? -eq 0 ]
    then
        echo "SQL Server Agent is running after $i attempts"
        break
    else
        echo "Attempt $i: SQL Server Agent not ready yet..."
        sleep 2
    fi
done

# Run the bootstrap script. We intentionally do NOT pass -b here: the bootstrap
# is not idempotent (re-running sp_cdc_enable_table on an already-CDC-enabled
# table raises Msg 22926), so -b combined with `restart: always` would crash-loop
# on any restart. Readiness is instead gated by the healthcheck, which verifies
# CDC is actually enabled on all fixture tables before the container is marked
# healthy (and thus before dependents like Terraform run).
echo "Running bootstrap script..."
$SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -i /docker-entrypoint-initdb.d/sqlserver_bootstrap.sql -C -t 30

if [ $? -eq 0 ]; then
    echo "Bootstrap script completed"
else
    echo "Warning: bootstrap script reported a non-zero exit; the healthcheck will gate readiness on CDC state"
fi

# Keep the container running for as long as sqlservr does. If it dies later the
# container exits non-zero and restart: always brings it back.
wait "$SQLSERVR_PID"
