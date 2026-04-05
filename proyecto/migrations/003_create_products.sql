CREATE TABLE IF NOT EXISTS products (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    price      NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    category   VARCHAR(50) NOT NULL CHECK (category IN ('electronics', 'clothing', 'food')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
