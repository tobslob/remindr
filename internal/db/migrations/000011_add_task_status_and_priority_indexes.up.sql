CREATE INDEX IF NOT EXISTS idx_tasks_active_user_status_created_id
ON tasks (user_id, status, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_active_user_priority_created_id
ON tasks (user_id, priority, created_at DESC, id DESC)
WHERE deleted_at IS NULL AND priority IS NOT NULL;
