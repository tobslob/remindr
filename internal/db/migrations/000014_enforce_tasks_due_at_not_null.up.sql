-- Backfill legacy tasks created before due_at became required so the
-- column can be enforced as NOT NULL going forward.
UPDATE tasks
SET due_at = COALESCE(created_at, updated_at, NOW())
WHERE due_at IS NULL;

-- Enforce due_at as NOT NULL to ensure all tasks have a due date going forward.
ALTER TABLE tasks
ALTER COLUMN due_at SET NOT NULL;
