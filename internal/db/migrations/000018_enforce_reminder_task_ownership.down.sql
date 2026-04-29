ALTER TABLE reminders
    DROP CONSTRAINT IF EXISTS reminders_task_owner_fkey;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_id_user_id_key;
