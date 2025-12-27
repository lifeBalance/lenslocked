CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE,
    token_hash TEXT UNIQUE NOT NULL
);

INSERT INTO sessions (user_id, token_hash)
VALUES ($1, $2)
RETURNING id;

UPDATE sessions
SET token_hash = $2
WHERE user_id = $1
RETURNING id;