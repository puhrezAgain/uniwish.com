CREATE TABLE products (
    id UUID PRIMARY KEY,
    store TEXT NOT NULL,
    store_product_id TEXT NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (store, store_product_id)
);

CREATE TABLE prices (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    price FLOAT NOT NULL,
    currency TEXT,
    scraped_at TIMESTAMPTZ NOT NULL DEFAULT now()
);