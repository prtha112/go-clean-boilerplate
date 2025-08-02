CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    item TEXT NOT NULL,
    amount INT NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    order_id TEXT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO orders (item, amount) VALUES ('Banana', 10);
INSERT INTO orders (item, amount) VALUES ('Apple', 5);
INSERT INTO orders (item, amount) VALUES ('Orange', 7);
INSERT INTO orders (item, amount) VALUES ('Grape', 12);
INSERT INTO orders (item, amount) VALUES ('Mango', 20);
INSERT INTO orders (item, amount) VALUES ('Pineapple', 15);
INSERT INTO orders (item, amount) VALUES ('Watermelon', 8);
INSERT INTO orders (item, amount) VALUES ('Strawberry', 25);
INSERT INTO orders (item, amount) VALUES ('Kiwi', 6);
INSERT INTO orders (item, amount) VALUES ('Papaya', 18);
INSERT INTO users (username, password) VALUES ('user1', '$2a$10$wzbcxhbeii1r7svJA51zCenSwbCnzld3Et54mXKv6tUku2AnOPt6S');
INSERT INTO users (username, password) VALUES ('user2', '$2a$10$hhMlzlc48p.nsLreP3nO1eoX8MUd1T8jr1YSYXTJ4Ap9oP9e5ky8.');
INSERT INTO users (username, password) VALUES ('user3', '$2a$10$G7SOtZ1w3RXURtSMfQI7huhvkcG4OP2XrN8rNOGmZabzSEHvHzABy');