-- (Don't use this one, the one at the bottom it's better)
-- CREATE TABLE IF NOT EXISTS sessions (
--     id SERIAL PRIMARY KEY,
--     user_id INT UNIQUE REFERENCES users (id),
--     token_hash TEXT UNIQUE NOT NULL
--  );

INSERT INTO sessions (user_id, token_hash)
VALUES ($1, $2)
RETURNING id;

UPDATE sessions
SET token_hash = $2
WHERE user_id = $1
RETURNING id;

SELECT user_id
FROM sessions
WHERE token_hash = $1;

SELECT email, password_hash
FROM users
WHERE id = $1;

DELETE FROM sessions
WHERE token_hash = $1;

-- How to add a foreign key to an existing table
ALTER TABLE sessions
    ADD CONSTRAINT sessions_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users;

-- How to delete the sessions for a user, when the latter is deleted.
-- when a user is deleted, so we don't have to:
-- 1. Delete its sessions first.
-- 2. Then delete the user.
-- We have to create the sessions table with the ON DELETE CASCADE on the FK
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL
 );
