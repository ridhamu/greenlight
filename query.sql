
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

