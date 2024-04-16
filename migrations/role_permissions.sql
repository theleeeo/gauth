CREATE TABLE IF NOT EXISTS role_permissions (
`role_id` VARCHAR(36) NOT NULL,
`p_key` VARCHAR(16) NOT NULL,
`p_val` VARCHAR(16) NOT NULL,
PRIMARY KEY (`role_id`, `p_key`, `p_val`),
FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE
);
