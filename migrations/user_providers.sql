CREATE TABLE IF NOT EXISTS user_providers (
`user_id` VARCHAR(36) NOT NULL,
-- What provider the user used to sign up
`provider` VARCHAR(10) NOT NULL,
-- The id of the user with the provider
`provider_id` VARCHAR(50) NOT NULL UNIQUE,
-- When was the provider added to the user
-- This will be the same as the user's `created_at` if it was the first provider
-- Otherwise, it will be the time the provider was added
`created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
);