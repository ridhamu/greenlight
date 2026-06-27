
-- `psql postgres`
SELECT current_user;

-- using super user, create database
CREATE DATABASE greenlight;

-- creating greenlight user
CREATE ROLE greenlight WITH LOGIN PASSWORD 'pa55word';

-- adding citext postgres extension
CREATE EXTENSION IF NOT EXISTS citext;

-- check all installed postgreq extension `\dx`


