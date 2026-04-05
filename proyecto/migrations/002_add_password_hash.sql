ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';

UPDATE users SET password_hash = '$2a$10$fQhqpTNLvajhdJvuSN27m.nls7rCu4wZ.EOVxWN1n4G3st7J/wYia';

ALTER TABLE users ALTER COLUMN password_hash DROP DEFAULT;
