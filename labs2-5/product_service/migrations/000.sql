CREATE TABLE IF NOT EXISTS products (
    id serial primary key ,
    name VARCHAR NOT NULL check(trim(name) <> ''),
    numOfItems int NOT NULL,
    PRICE numeric(100, 2) NOT NULL
);

SELECT id FROM products;