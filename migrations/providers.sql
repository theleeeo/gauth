CREATE TABLE IF NOT EXISTS user_providers (
`user_id` VARCHAR(50) NOT NULL,
-- What provider the user used to sign up
`provider` VARCHAR(10) NOT NULL,
-- The id of the user with the provider
`provider_id` VARCHAR(50) NOT NULL UNIQUE,
FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
);