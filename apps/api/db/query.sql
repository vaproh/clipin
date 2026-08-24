-- name: GetMigrationVersion :one
SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1;
