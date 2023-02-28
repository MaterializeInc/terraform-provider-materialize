CREATE TABLE table1 (
    id INT GENERATED ALWAYS AS IDENTITY,
);

CREATE TABLE table2 (
    id INT,
    updated_at timestamp NOT NULL
);

-- Enable REPLICA for both tables
ALTER TABLE table1 REPLICA IDENTITY FULL;
ALTER TABLE table2 REPLICA IDENTITY FULL;

-- Create publication on the created tables
CREATE PUBLICATION mz_source FOR TABLE table1, table2;

-- Create user and role to be used by Materialize
CREATE ROLE materialize REPLICATION LOGIN PASSWORD 'c2VjcmV0Cg==';
GRANT SELECT ON table1, table2 TO materialize;
