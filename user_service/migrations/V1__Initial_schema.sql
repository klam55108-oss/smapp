CREATE TABLE users (
    id BINARY(16) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    handle VARCHAR(20) NOT NULL,
    email VARCHAR(254) NOT NULL,
    image_url VARCHAR(2000),
    password_hash VARCHAR(60) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY email_unique (email),
    UNIQUE KEY handle_unique (handle)
);