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
# sp_cdc_enable_table fails with 14258 ("SQL Server Agent is starting") until it
# has. sys.dm_server_services reports Running before jobs can be added, so the
# probe is the operation itself: add and drop a throwaway job.
echo "Waiting for SQL Server Agent to accept jobs..."
for i in {1..90}
do
    $SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -C -b -l 5 -Q "
        SET NOCOUNT ON;
        IF EXISTS (SELECT 1 FROM msdb.dbo.sysjobs WHERE name = N'mz_fixture_agent_probe')
            EXEC msdb.dbo.sp_delete_job @job_name = N'mz_fixture_agent_probe';
        EXEC msdb.dbo.sp_add_job @job_name = N'mz_fixture_agent_probe';
        EXEC msdb.dbo.sp_delete_job @job_name = N'mz_fixture_agent_probe';" > /dev/null 2>&1
    if [ $? -eq 0 ]
    then
        echo "SQL Server Agent accepted a job after $i attempts"
        break
    else
        echo "Attempt $i: SQL Server Agent not ready yet..."
        sleep 2
    fi
done

# The bootstrap creates the database and tables, enables CDC and seeds data. It
# runs once and without -b: the seed INSERTs are not idempotent, and a single
# failed statement must not abort the rest.
echo "Running bootstrap script..."
$SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -i /docker-entrypoint-initdb.d/sqlserver_bootstrap.sql -C -t 30

if [ $? -eq 0 ]; then
    echo "Bootstrap script completed"
else
    echo "Warning: bootstrap script reported a non-zero exit; repairing CDC state below"
fi

# The healthcheck needs CDC on every fixture table. If any sp_cdc_enable_table
# call failed above, enable CDC on whatever is still untracked until all of them
# are, rather than leaving the container unhealthy for the rest of the job.
echo "Verifying CDC on the fixture tables..."
for i in {1..30}
do
    tracked=$($SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -C -h -1 -W -Q "SET NOCOUNT ON; SELECT COUNT(*) FROM testdb.sys.tables WHERE is_tracked_by_cdc = 1;" 2>/dev/null | tr -dc '0-9')
    if [ "${tracked:-0}" -ge 10 ]
    then
        echo "CDC is enabled on $tracked fixture tables"
        break
    fi
    echo "Attempt $i: CDC enabled on ${tracked:-0}/10 tables, enabling the rest..."
    $SQLCMD -S localhost -U sa -P "${SA_PASSWORD}" -C -d testdb -Q "
        SET NOCOUNT ON;
        IF (SELECT is_cdc_enabled FROM sys.databases WHERE name = N'testdb') = 0
            EXEC sys.sp_cdc_enable_db;
        DECLARE @t sysname;
        DECLARE untracked CURSOR LOCAL FAST_FORWARD FOR
            SELECT name FROM sys.tables
            WHERE schema_id = SCHEMA_ID(N'dbo') AND is_tracked_by_cdc = 0 AND name LIKE N'table%';
        OPEN untracked;
        FETCH NEXT FROM untracked INTO @t;
        WHILE @@FETCH_STATUS = 0
        BEGIN
            BEGIN TRY
                EXEC sys.sp_cdc_enable_table @source_schema = N'dbo', @source_name = @t, @role_name = NULL, @supports_net_changes = 0;
            END TRY
            BEGIN CATCH
                PRINT 'CDC enable failed for ' + @t + ': ' + ERROR_MESSAGE();
            END CATCH
            FETCH NEXT FROM untracked INTO @t;
        END
        CLOSE untracked;
        DEALLOCATE untracked;" 2>&1 | grep -v '^\s*$'
    sleep 5
done

# Keep the container running for as long as sqlservr does. If it dies later the
# container exits non-zero and restart: always brings it back.
wait "$SQLSERVR_PID"
