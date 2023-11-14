CREATE TABLE IF NOT EXISTS products (
    code VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    numOfItems NUMERIC NOT NULL,
    PRICE VARCHAR NOT NULL
)

SELECT code FROM products

INSERT INTO products VALUES ('test', 'test', 5, 'test')

DELETE FROM products;