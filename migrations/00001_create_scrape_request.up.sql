
CREATE TABLE scrape_requests (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPZ NOT NULL DEFAULT now()
);