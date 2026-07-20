
-- `psql postgres`
SELECT current_user;

-- using super user, create database
CREATE DATABASE greenlight;

-- creating greenlight user
CREATE ROLE greenlight WITH LOGIN PASSWORD 'pa55word';

-- adding citext postgres extension
CREATE EXTENSION IF NOT EXISTS citext;

-- check all installed postgreq extension `\dx`

-- insert into movies
INSERT INTO movies (title, year, runtime, genres)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version

-- Fetching a Movie
SELECT id, created_at, title, year, runtime, genres, version FROM movies WHERE id = $1

-- update a movie
UPDATE movies
SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
WHERE id = $5
RETURNING version

-- delete a movie
DELETE FROM movies
WHERE id = $1

--update movie(optimistic locking)
UPDATE movies
SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
WHERE id = $5 AND version = $6
RETURNING version

-- GetAll query
SELECT id, created_at, title, year, runtime, genres, version
FROM movies
ORDER BY id


-- Insert user
INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version

-- insert token
INSERT INTO tokens (hash, user_id, expiry, scope) VALUES ($1, $2, $3, $4)

-- delete token based on userID and context
DELETE FROM tokens WHERE scope = $1 AND user_id = $2;

-- getting user info based on token
SELECT
	users.id,
	users.created_at,
	users.name,
	users.email,
	users.password_hash,
	users.activated,
	users.version
FROM
	users
	INNER JOIN tokens ON users.id = tokens.user_id
WHERE
  tokens.hash = $1
	AND tokens.scope = $2
	AND tokens.expiry > $3


-- add user permissions
INSERT INTO
	users_permissions
SELECT
	$1,
	permissions.id
FROM
	permissions
WHERE
	permissions.code = ANY()


-- look user's permissions
SELECT
	email,
	code
FROM
	users
	INNER JOIN users_permissions ON users.id = users_permissions.user_id
	INNER JOIN permissions ON users_permissions.permission_id = permissions.id
WHERE
	users.email = 'trump@donald.com';
