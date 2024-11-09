CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(27) NOT NULL,
    account_id VARCHAR(27) NOT NULL,
    price MONEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS order_products (
    order_id VARCHAR(27) REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(27) NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (order_id, product_id)
);