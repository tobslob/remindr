CREATE INDEX IF NOT EXISTS idx_tasks_active_user_created_id
ON tasks (user_id, created_at DESC, id DESC)
WHERE deleted_at IS NULL;
