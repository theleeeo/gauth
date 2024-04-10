CREATE TABLE IF NOT EXISTS users (
`id` VARCHAR(50) NOT NULL PRIMARY KEY,
`email` VARCHAR(50) NOT NULL,
`first_name` VARCHAR(50) NOT NULL,
`last_name` VARCHAR(50) NOT NULL,
`role` ENUM('user', 'admin') NOT NULL DEFAULT 'user'
);