CREATE TABLE urls (
    id SERIAL,
    url TEXT NOT NULL,
    shortcode TEXT UNIQUE NOT NULL,
    createdat TIMESTAMP NOT NULL,
    updatedat TIMESTAMP NOT NULL,
    accesscount INT NOT NULL DEFAULT 0,
    ownerid INT NOT NULL
    );