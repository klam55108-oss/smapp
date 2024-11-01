CREATE TABLE posts (
    id BINARY(16) PRIMARY KEY,
    body TEXT,
    author_id BINARY(16) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE likes (
    id BINARY(16) PRIMARY KEY,
    liked_id BINARY(16) NOT NULL,
    liked_type ENUM('post', 'comment') NOT NULL,
    author_id BINARY(16) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    author_id BINARY(16) NOT NULL,
    body TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE likes_count (
    id BINARY(16) PRIMARY KEY,
    liked_id BINARY(16) NOT NULL,
    liked_type ENUM('post', 'comment') NOT NULL,
    count INT SIGNED NOT NULL DEFAULT 0
);

CREATE TABLE comments_count (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    count INT SIGNED NOT NULL DEFAULT 0
);

CREATE TABLE images (
    id BINARY(16) PRIMARY KEY,
    post_id BINARY(16) NOT NULL,
    position INT SIGNED NOT NULL,
    s3_bucket VARCHAR(63) NOT NULL,
    s3_key VARCHAR(1024) NOT NULL
);