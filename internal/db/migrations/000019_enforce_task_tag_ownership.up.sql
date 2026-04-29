ALTER TABLE tags
    ADD CONSTRAINT tags_id_user_id_key UNIQUE (id, user_id);

ALTER TABLE task_tags
    ADD COLUMN user_id UUID;

UPDATE task_tags tt
SET user_id = tk.user_id
FROM tasks tk, tags tg
WHERE tk.id = tt.task_id
  AND tg.id = tt.tag_id
  AND tk.user_id = tg.user_id;

DO $$
DECLARE
    invalid_count BIGINT;
BEGIN
    SELECT COUNT(*)
    INTO invalid_count
    FROM task_tags
    WHERE user_id IS NULL;

    IF invalid_count > 0 THEN
        RAISE EXCEPTION 'cannot enforce task_tags owner integrity: % row(s) have mismatched task/tag owners', invalid_count;
    END IF;
END $$;

ALTER TABLE task_tags
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE task_tags
    ADD CONSTRAINT task_tags_task_owner_fkey
        FOREIGN KEY (task_id, user_id) REFERENCES tasks(id, user_id) ON DELETE CASCADE;

ALTER TABLE task_tags
    ADD CONSTRAINT task_tags_tag_owner_fkey
        FOREIGN KEY (tag_id, user_id) REFERENCES tags(id, user_id) ON DELETE CASCADE;

CREATE INDEX idx_task_tags_tag_user_id ON task_tags(tag_id, user_id);
