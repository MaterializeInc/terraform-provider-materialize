CREATE DATABASE IF NOT EXISTS shop;
USE shop;

GRANT ALL PRIVILEGES ON shop.* TO 'mysqluser';

CREATE USER 'repluser'@'%' IDENTIFIED WITH mysql_native_password BY 'c2VjcmV0Cg==';

GRANT SELECT, RELOAD, SHOW DATABASES, REPLICATION SLAVE, REPLICATION CLIENT, LOCK TABLES ON *.* TO 'repluser'@'%';

FLUSH PRIVILEGES;

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
    banned BOOLEAN
    -- TODO: Disable until https://github.com/MaterializeInc/materialize/issues/24952 is resolved
    -- updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS mysql_table3
(
    id INT AUTO_INCREMENT PRIMARY KEY
);

-- Insert sample data
INSERT INTO mysql_table1 (id) VALUES (NULL), (NULL), (NULL), (NULL), (NULL);
-- INSERT INTO mysql_table2 (id, updated_at) VALUES (1, NOW()), (2, NOW()), (3, NOW()), (4, NOW()), (5, NOW());
INSERT INTO mysql_table2 (id) VALUES (1), (2), (3), (4), (5);
INSERT INTO mysql_table3 (id) VALUES (NULL), (NULL), (NULL), (NULL), (NULL);
