FROM mcr.microsoft.com/mssql/server:2022-latest

# Switch to root to install packages and copy files
USER root

# Add SQL Server tools to PATH (they should already be installed in the base image)
ENV PATH="$PATH:/opt/mssql-tools/bin:/opt/mssql-tools18/bin"

ENV MSSQL_AGENT_ENABLED=true
ENV MSSQL_TCP_PORT=1433

# Copy bootstrap files
COPY sqlserver_bootstrap.sql /docker-entrypoint-initdb.d/
COPY entrypoint.sh /entrypoint.sh

# Make script executable
RUN chmod +x /entrypoint.sh

# Create directory for SQL Server to avoid permission issues
RUN mkdir -p /var/opt/mssql && chown -R mssql:root /var/opt/mssql

# Set user back to mssql
USER mssql

# Use custom entrypoint
ENTRYPOINT ["/entrypoint.sh"]
