CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

-- Get the list of logged in users.
-- Since INNER is the most common, you can omit it, and use just JOIN.
SELECT * FROM users
    INNER JOIN sessions ON users.id = sessions.user_id;

SELECT * FROM users
    LEFT JOIN sessions ON users.id = sessions.user_id;

SELECT * FROM users
    RIGHT JOIN sessions ON users.id = sessions.user_id;

SELECT * FROM users
    FULL OUTER JOIN sessions ON users.id = sessions.user_id;

-- Get the list of logged in users (but selecting some fields from both tables).
SELECT users.id, users.email, sessions.token_hash FROM users
    INNER JOIN sessions ON users.id = sessions.user_id;