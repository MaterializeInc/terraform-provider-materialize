ALTER SYSTEM SET wal_level = logical;
ALTER ROLE postgres WITH REPLICATION;

CREATE TABLE table1 (
    id INT GENERATED ALWAYS AS IDENTITY
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

INSERT INTO table1 VALUES (1), (2), (3), (4), (5);
INSERT INTO table2 VALUES (1, NOW()), (2, NOW()), (3, NOW()), (4, NOW()), (5, NOW());
