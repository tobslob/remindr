ALTER TABLE tasks
    ADD CONSTRAINT tasks_id_user_id_key UNIQUE (id, user_id);

ALTER TABLE reminders
    ADD CONSTRAINT reminders_task_owner_fkey
        FOREIGN KEY (task_id, user_id) REFERENCES tasks(id, user_id) ON DELETE CASCADE;
