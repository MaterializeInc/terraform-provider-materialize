CREATE DATABASE IF NOT EXISTS shop;
USE shop;

GRANT ALL PRIVILEGES ON shop.* TO 'mysqluser';

CREATE USER 'repluser'@'%' IDENTIFIED WITH mysql_native_password BY 'c2VjcmV0Cg==';

GRANT SELECT, RELOAD, SHOW DATABASES, REPLICATION SLAVE, REPLICATION CLIENT, LOCK TABLES ON *.* TO 'repluser'@'%';

FLUSH PRIVILEGES;

-- Basic tables
CREATE TABLE IF NOT EXISTS mysql_table1
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    about TEXT,
    banned BOOLEAN
);

CREATE TABLE IF NOT EXISTS mysql_table2
(
    id INT,
    name VARCHAR(255),
    about TEXT,
    banned BOOLEAN,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS mysql_table3
(
    id INT AUTO_INCREMENT PRIMARY KEY
);

-- Table with ENUM type (needs text_columns)
CREATE TABLE IF NOT EXISTS mysql_table4
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    status ENUM('active', 'inactive', 'deleted') NOT NULL DEFAULT 'active'
);

-- Table with various data types
CREATE TABLE IF NOT EXISTS mysql_table5
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255),
    age INT,
    salary DECIMAL(10, 2),
    is_active BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Table with SET type (needs text_columns)
CREATE TABLE IF NOT EXISTS mysql_table6
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    tags SET('tag1', 'tag2', 'tag3', 'tag4') DEFAULT NULL,
    status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending'
);

-- Table with special characters in names
CREATE TABLE IF NOT EXISTS mysql_table7
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    `name_with_underscores` VARCHAR(255),
    `name-with-dashes` VARCHAR(255),
    description TEXT,
    binary_data BLOB
);

-- Table with nullable columns
CREATE TABLE IF NOT EXISTS mysql_table8
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    nullable_string VARCHAR(255) NULL,
    nullable_int INT NULL,
    nullable_timestamp TIMESTAMP NULL,
    required_field VARCHAR(255) NOT NULL
);

-- Table with numeric types
CREATE TABLE IF NOT EXISTS mysql_table9
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    tiny_int TINYINT,
    small_int SMALLINT,
    medium_int MEDIUMINT,
    big_int BIGINT,
    float_val FLOAT,
    double_val DOUBLE,
    decimal_val DECIMAL(10, 2)
);

-- Table with date/time types
CREATE TABLE IF NOT EXISTS mysql_table10
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    date_col DATE,
    time_col TIME,
    datetime_col DATETIME,
    timestamp_col TIMESTAMP,
    year_col YEAR
);

-- Insert sample data
INSERT INTO mysql_table1 (name, about, banned) VALUES 
    ('John Doe', 'Lorem ipsum dolor sit amet', 0), 
    ('Jane Doe', 'Lorem ipsum dolor sit amet', 1), 
    ('Alice', 'Lorem ipsum dolor sit amet', 0), 
    ('Bob', 'Lorem ipsum dolor sit amet', 1), 
    ('Charlie', 'Lorem ipsum dolor sit amet', 0);

INSERT INTO mysql_table2 (id, name, about, banned, updated_at) VALUES 
    (1, 'Record 1', 'First record', 0, NOW()), 
    (2, 'Record 2', 'Second record', 1, NOW()), 
    (3, 'Record 3', 'Third record', 0, NOW()), 
    (4, 'Record 4', 'Fourth record', 1, NOW()), 
    (5, 'Record 5', 'Fifth record', 0, NOW());

INSERT INTO mysql_table3 (id) VALUES (NULL), (NULL), (NULL), (NULL), (NULL);

INSERT INTO mysql_table4 (status) VALUES 
    ('active'), ('inactive'), ('deleted'), ('active'), ('inactive');

INSERT INTO mysql_table5 (name, email, age, salary, is_active, created_at) VALUES
    ('John Doe', 'john@example.com', 30, 50000.00, 1, NOW()),
    ('Jane Smith', 'jane@example.com', 25, 60000.00, 1, NOW()),
    ('Bob Johnson', 'bob@example.com', 35, 55000.00, 0, NOW()),
    ('Alice Brown', 'alice@example.com', 28, 65000.00, 1, NOW()),
    ('Charlie Wilson', 'charlie@example.com', 32, 70000.00, 1, NOW());

INSERT INTO mysql_table6 (tags, status) VALUES
    ('tag1,tag2', 'pending'),
    ('tag3', 'approved'),
    ('tag1,tag3,tag4', 'rejected'),
    (NULL, 'pending'),
    ('tag2', 'approved');

INSERT INTO mysql_table7 (`name_with_underscores`, `name-with-dashes`, description, binary_data) VALUES
    ('underscore_name', 'dash-name', 'Description with special chars: !@#$%', UNHEX('DEADBEEF')),
    ('another_underscore', 'another-dash', 'More special chars: &*()', UNHEX('CAFEBABE'));

INSERT INTO mysql_table8 (nullable_string, nullable_int, nullable_timestamp, required_field) VALUES
    ('Has value', 100, NOW(), 'Required'),
    (NULL, NULL, NULL, 'Required'),
    ('Another value', 200, NOW(), 'Required'),
    (NULL, 300, NULL, 'Required');

INSERT INTO mysql_table9 (tiny_int, small_int, medium_int, big_int, float_val, double_val, decimal_val) VALUES
    (127, 32767, 8388607, 9223372036854775807, 3.14159, 2.71828, 1234.56),
    (-128, -32768, -8388608, -9223372036854775808, -3.14159, -2.71828, -1234.56),
    (0, 0, 0, 0, 0.0, 0.0, 0.00);

INSERT INTO mysql_table10 (date_col, time_col, datetime_col, timestamp_col, year_col) VALUES
    ('2024-01-01', '12:00:00', '2024-01-01 12:00:00', NOW(), 2024),
    ('2024-06-15', '18:30:00', '2024-06-15 18:30:00', NOW(), 2024),
    ('2024-12-31', '23:59:59', '2024-12-31 23:59:59', NOW(), 2024);
