DROP INDEX IF EXISTS idx_task_tags_tag_user_id;

ALTER TABLE task_tags
    DROP CONSTRAINT IF EXISTS task_tags_tag_owner_fkey;

ALTER TABLE task_tags
    DROP CONSTRAINT IF EXISTS task_tags_task_owner_fkey;

ALTER TABLE task_tags
    DROP COLUMN IF EXISTS user_id;

ALTER TABLE tags
    DROP CONSTRAINT IF EXISTS tags_id_user_id_key;
