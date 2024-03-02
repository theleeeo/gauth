CREATE TABLE IF NOT EXISTS user (
`id` string NOT NULL PRIMARY KEY
`nickname` VARCHAR(50)
`role` ENUM('user', 'admin') NOT NULL DEFAULT 'user'
-- What provider the user used to sign up
`provider` ENUM('github') NOT NULL
-- The id of the user with the provider
`provider_id` VARCHAR(50)
);