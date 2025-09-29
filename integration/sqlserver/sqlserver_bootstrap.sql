-- Create the test database
IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'testdb')
BEGIN
    CREATE DATABASE testdb;
END
GO

-- Enable snapshot isolation settings required for CDC with Materialize
ALTER DATABASE testdb SET ALLOW_SNAPSHOT_ISOLATION ON;
GO

ALTER DATABASE testdb SET READ_COMMITTED_SNAPSHOT ON;
GO

USE testdb;
GO

-- Enable Change Data Capture on the database
EXEC sys.sp_cdc_enable_db;
GO

-- Create test tables
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table1]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table1] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255),
        about NTEXT,
        banned BIT,
        created_at DATETIME2 DEFAULT GETDATE()
    );
END
GO

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table2]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table2] (
        id INT PRIMARY KEY,
        name NVARCHAR(255),
        about NVARCHAR(255),
        banned BIT,
        updated_at DATETIME2 NOT NULL DEFAULT GETDATE()
    );
END
GO

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table3]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table3] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        status NVARCHAR(50) NOT NULL DEFAULT 'active',
        data XML,
        created_at DATETIME2 DEFAULT GETDATE()
    );
END
GO

-- Enable CDC on the tables
EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table1',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table2',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table3',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

-- Insert sample data
INSERT INTO [dbo].[table1] (name, about, banned) 
VALUES 
    ('John Doe', 'Lorem ipsum dolor sit amet', 0),
    ('Jane Doe', 'Lorem ipsum dolor sit amet', 1),
    ('Alice Smith', 'Lorem ipsum dolor sit amet', 0),
    ('Bob Johnson', 'Lorem ipsum dolor sit amet', 1),
    ('Charlie Brown', 'Lorem ipsum dolor sit amet', 0);
GO

INSERT INTO [dbo].[table2] (id, name, about, banned, updated_at) 
VALUES 
    (1, 'Record 1', 'First record', 0, GETDATE()),
    (2, 'Record 2', 'Second record', 1, GETDATE()),
    (3, 'Record 3', 'Third record', 0, GETDATE()),
    (4, 'Record 4', 'Fourth record', 1, GETDATE()),
    (5, 'Record 5', 'Fifth record', 0, GETDATE());
GO

INSERT INTO [dbo].[table3] (status, data)
VALUES 
    ('active', '<data>Sample XML 1</data>'),
    ('inactive', '<data>Sample XML 2</data>'),
    ('active', '<data>Sample XML 3</data>'),
    ('deleted', '<data>Sample XML 4</data>'),
    ('active', '<data>Sample XML 5</data>');
GO

-- Verify snapshot isolation settings
SELECT 
    name,
    is_cdc_enabled,
    snapshot_isolation_state,
    snapshot_isolation_state_desc,
    is_read_committed_snapshot_on
FROM sys.databases 
WHERE name = 'testdb';
GO

-- Verify CDC is enabled
SELECT name, is_cdc_enabled FROM sys.databases WHERE name = 'testdb';
GO

SELECT 
    s.name AS schema_name,
    t.name AS table_name,
    t.is_tracked_by_cdc
FROM sys.tables t
INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
WHERE t.is_tracked_by_cdc = 1;
GO 
