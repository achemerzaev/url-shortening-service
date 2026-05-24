CREATE TABLE url_users (
    id SERIAL PRIMARY KEY, 
    name TEXT NOT NULL, 
    email TEXT UNIQUE NOT NULL, 
    password TEXT NOT NULL
    );