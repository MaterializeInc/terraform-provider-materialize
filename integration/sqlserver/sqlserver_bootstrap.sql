-- Create the test database (if it doesn't exist)
IF DB_ID('testdb') IS NULL
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

-- Basic tables
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table1]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table1] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255),
        about NTEXT,  -- Unsupported type, needs exclude_columns
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

-- Table with various data types
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table4]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table4] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255),
        email NVARCHAR(255),
        age INT,
        salary DECIMAL(10, 2),
        is_active BIT,
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2
    );
END
GO

-- Table with unsupported types (needs exclude_columns or text_columns)
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table5]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table5] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255),
        large_text NTEXT,  -- Unsupported, needs exclude_columns
        image_data IMAGE,  -- Unsupported, needs exclude_columns
        xml_data XML,
        json_data NVARCHAR(MAX)
    );
END
GO

-- Table with special characters in column names
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table6]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table6] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        [name_with_underscores] NVARCHAR(255),
        [name-with-dashes] NVARCHAR(255),
        [name with spaces] NVARCHAR(255),
        description NVARCHAR(MAX),
        binary_data VARBINARY(MAX)
    );
END
GO

-- Table with nullable columns
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table7]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table7] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        nullable_string NVARCHAR(255) NULL,
        nullable_int INT NULL,
        nullable_timestamp DATETIME2 NULL,
        required_field NVARCHAR(255) NOT NULL
    );
END
GO

-- Table with numeric types
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table8]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table8] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        tiny_int TINYINT,
        small_int SMALLINT,
        big_int BIGINT,
        float_val FLOAT,
        real_val REAL,
        decimal_val DECIMAL(10, 2),
        money_val MONEY,
        smallmoney_val SMALLMONEY
    );
END
GO

-- Table with date/time types
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table9]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table9] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        date_col DATE,
        time_col TIME,
        datetime_col DATETIME,
        datetime2_col DATETIME2,
        datetimeoffset_col DATETIMEOFFSET,
        smalldatetime_col SMALLDATETIME
    );
END
GO

-- Table with text types (some unsupported)
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[table10]') AND type in (N'U'))
BEGIN
    CREATE TABLE [dbo].[table10] (
        id INT IDENTITY(1,1) PRIMARY KEY,
        varchar_col NVARCHAR(255),
        char_col NCHAR(10),
        text_col NTEXT,  -- Unsupported, needs exclude_columns
        nvarchar_max NVARCHAR(MAX)
    );
END
GO

-- Enable CDC on all tables
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

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table4',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table5',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table6',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table7',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table8',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table9',
    @role_name = NULL,
    @supports_net_changes = 0;
GO

EXEC sys.sp_cdc_enable_table
    @source_schema = N'dbo',
    @source_name = N'table10',
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

INSERT INTO [dbo].[table4] (name, email, age, salary, is_active, created_at) 
VALUES
    ('John Doe', 'john@example.com', 30, 50000.00, 1, GETDATE()),
    ('Jane Smith', 'jane@example.com', 25, 60000.00, 1, GETDATE()),
    ('Bob Johnson', 'bob@example.com', 35, 55000.00, 0, GETDATE()),
    ('Alice Brown', 'alice@example.com', 28, 65000.00, 1, GETDATE()),
    ('Charlie Wilson', 'charlie@example.com', 32, 70000.00, 1, GETDATE());
GO

INSERT INTO [dbo].[table5] (name, large_text, image_data, xml_data, json_data)
VALUES
    ('Table 5 Record 1', 'Large text content here', NULL, '<data>XML 1</data>', '{"key": "value1"}'),
    ('Table 5 Record 2', 'More large text', NULL, '<data>XML 2</data>', '{"key": "value2"}'),
    ('Table 5 Record 3', 'Even more text', NULL, NULL, '{"key": "value3"}');
GO

INSERT INTO [dbo].[table6] ([name_with_underscores], [name-with-dashes], [name with spaces], description, binary_data)
VALUES
    ('underscore_name', 'dash-name', 'space name', 'Description with special chars: !@#$%', 0xDEADBEEF),
    ('another_underscore', 'another-dash', 'another space', 'More special chars: &*()', 0xCAFEBABE);
GO

INSERT INTO [dbo].[table7] (nullable_string, nullable_int, nullable_timestamp, required_field)
VALUES
    ('Has value', 100, GETDATE(), 'Required'),
    (NULL, NULL, NULL, 'Required'),
    ('Another value', 200, GETDATE(), 'Required'),
    (NULL, 300, NULL, 'Required');
GO

INSERT INTO [dbo].[table8] (tiny_int, small_int, big_int, float_val, real_val, decimal_val, money_val, smallmoney_val)
VALUES
    (255, 32767, 9223372036854775807, 3.14159, 2.71828, 1234.56, 999999.99, 214748.3647),
    (0, -32768, -9223372036854775808, -3.14159, -2.71828, -1234.56, -999999.99, -214748.3647),
    (128, 0, 0, 0.0, 0.0, 0.00, 0.00, 0.00);
GO

INSERT INTO [dbo].[table9] (date_col, time_col, datetime_col, datetime2_col, datetimeoffset_col, smalldatetime_col)
VALUES
    ('2024-01-01', '12:00:00', '2024-01-01 12:00:00', '2024-01-01 12:00:00.1234567', '2024-01-01 12:00:00 +00:00', '2024-01-01 12:00:00'),
    ('2024-06-15', '18:30:00', '2024-06-15 18:30:00', '2024-06-15 18:30:00.1234567', '2024-06-15 18:30:00 +00:00', '2024-06-15 18:30:00'),
    ('2024-12-31', '23:59:59', '2024-12-31 23:59:59', '2024-12-31 23:59:59.1234567', '2024-12-31 23:59:59 +00:00', '2024-12-31 23:59:59');
GO

INSERT INTO [dbo].[table10] (varchar_col, char_col, text_col, nvarchar_max)
VALUES
    ('Varchar value 1', 'Char1     ', 'Large text content 1', 'NVARCHAR(MAX) content 1'),
    ('Varchar value 2', 'Char2     ', 'Large text content 2', 'NVARCHAR(MAX) content 2'),
    ('Varchar value 3', 'Char3     ', NULL, 'NVARCHAR(MAX) content 3');
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
WHERE t.is_tracked_by_cdc = 1
ORDER BY t.name;
GO 
