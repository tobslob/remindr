CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_tasks_active_search_trgm
ON tasks
USING GIN ((title || ' ' || description) gin_trgm_ops)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_active_user_due_at
ON tasks (user_id, due_at)
WHERE deleted_at IS NULL AND due_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_active_user_completed_at
ON tasks (user_id, completed_at)
WHERE deleted_at IS NULL AND completed_at IS NOT NULL;
