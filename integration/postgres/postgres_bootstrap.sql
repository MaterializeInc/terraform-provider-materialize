ALTER SYSTEM SET wal_level = logical;
ALTER ROLE postgres WITH REPLICATION;

-- Basic tables
CREATE TABLE table1 (
    id INT GENERATED ALWAYS AS IDENTITY
);

CREATE TABLE table2 (
    id INT,
    updated_at timestamp NOT NULL
);

CREATE TABLE table3 (
    id INT GENERATED ALWAYS AS IDENTITY
);

-- Table with various data types for testing
CREATE TABLE table4 (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255),
    age INT,
    salary DECIMAL(10, 2),
    is_active BOOLEAN,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Table with unsupported types that need text_columns
CREATE TABLE table5 (
    id INT PRIMARY KEY,
    data JSONB,
    tags TEXT[],
    metadata JSONB
);

-- Table with special characters and edge cases
CREATE TABLE table6 (
    id INT PRIMARY KEY,
    name_with_dots VARCHAR(255),
    "name-with-dashes" VARCHAR(255),
    "name with spaces" VARCHAR(255),
    description TEXT,
    binary_data BYTEA
);

-- Table with nullable columns
CREATE TABLE table7 (
    id INT PRIMARY KEY,
    nullable_string VARCHAR(255),
    nullable_int INT,
    nullable_timestamp TIMESTAMP,
    required_field VARCHAR(255) NOT NULL
);

-- Table with numeric types
CREATE TABLE table8 (
    id SERIAL PRIMARY KEY,
    small_int SMALLINT,
    big_int BIGINT,
    real_num REAL,
    double_num DOUBLE PRECISION,
    numeric_val NUMERIC(10, 2)
);

-- Enable REPLICA IDENTITY for all tables
ALTER TABLE table1 REPLICA IDENTITY FULL;
ALTER TABLE table2 REPLICA IDENTITY FULL;
ALTER TABLE table3 REPLICA IDENTITY FULL;
ALTER TABLE table4 REPLICA IDENTITY FULL;
ALTER TABLE table5 REPLICA IDENTITY FULL;
ALTER TABLE table6 REPLICA IDENTITY FULL;
ALTER TABLE table7 REPLICA IDENTITY FULL;
ALTER TABLE table8 REPLICA IDENTITY FULL;

-- Create publication on all tables
CREATE PUBLICATION mz_source FOR TABLE table1, table2, table3, table4, table5, table6, table7, table8;

-- Insert sample data
INSERT INTO table1 VALUES (1), (2), (3), (4), (5);
INSERT INTO table2 VALUES (1, NOW()), (2, NOW()), (3, NOW()), (4, NOW()), (5, NOW());
INSERT INTO table3 VALUES (1), (2), (3), (4), (5);

INSERT INTO table4 (id, name, email, age, salary, is_active, created_at) VALUES
    (1, 'John Doe', 'john@example.com', 30, 50000.00, true, NOW()),
    (2, 'Jane Smith', 'jane@example.com', 25, 60000.00, true, NOW()),
    (3, 'Bob Johnson', 'bob@example.com', 35, 55000.00, false, NOW()),
    (4, 'Alice Brown', 'alice@example.com', 28, 65000.00, true, NOW()),
    (5, 'Charlie Wilson', 'charlie@example.com', 32, 70000.00, true, NOW());

INSERT INTO table5 (id, data, tags, metadata) VALUES
    (1, '{"key": "value", "number": 123}'::jsonb, ARRAY['tag1', 'tag2'], '{"meta": "data"}'::jsonb),
    (2, '{"key": "value2", "number": 456}'::jsonb, ARRAY['tag3'], '{"meta": "data2"}'::jsonb),
    (3, '{"key": "value3"}'::jsonb, ARRAY['tag1', 'tag3'], NULL::jsonb);

INSERT INTO table6 (id, name_with_dots, "name-with-dashes", "name with spaces", description, binary_data) VALUES
    (1, 'table.name', 'table-name', 'table name', 'Description with special chars: !@#$%', E'\\xDEADBEEF'),
    (2, 'another.table', 'another-name', 'another name', 'More special chars: &*()', E'\\xCAFEBABE');

INSERT INTO table7 (id, nullable_string, nullable_int, nullable_timestamp, required_field) VALUES
    (1, 'Has value', 100, NOW(), 'Required'),
    (2, NULL, NULL, NULL, 'Required'),
    (3, 'Another value', 200, NOW(), 'Required'),
    (4, NULL, 300, NULL, 'Required');

INSERT INTO table8 (small_int, big_int, real_num, double_num, numeric_val) VALUES
    (32767, 9223372036854775807, 3.14159, 2.71828, 1234.56),
    (-32768, -9223372036854775808, -3.14159, -2.71828, -1234.56),
    (0, 0, 0.0, 0.0, 0.00);
