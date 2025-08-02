CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    item TEXT NOT NULL,
    amount INT NOT NULL
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