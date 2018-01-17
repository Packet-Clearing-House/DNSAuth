-- Opening the CLI
psql -h localhost -p 5432 -d pipeline


-- Creating the customer table
DROP TABLE ns_customers;
CREATE TABLE ns_customers(
   ip TEXT PRIMARY KEY NOT NULL,
   name TEXT,
   asn BOOL,
   prefix BOOL
);


-- Inserting some customers..
INSERT INTO ns_customers VALUES ('1.199.71.00/24', 'Foo', true, true);
INSERT INTO ns_customers VALUES ('caec:cec6:c4ef:bb7b::/48', 'Bar', true, true);
INSERT INTO ns_customers VALUES ('11.206.206.0/24', 'Bash', true, true);


