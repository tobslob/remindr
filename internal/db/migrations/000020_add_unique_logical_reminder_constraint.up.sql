DO $$
DECLARE
    duplicate_slot_count BIGINT;
BEGIN
    SELECT COUNT(*)
    INTO duplicate_slot_count
    FROM (
        SELECT 1
        FROM reminders
        GROUP BY task_id, user_id, type, remind_at
        HAVING COUNT(*) > 1
    ) duplicate_slots;

    IF duplicate_slot_count > 0 THEN
        RAISE EXCEPTION 'cannot enforce unique logical reminders: % duplicate reminder slot(s) exist', duplicate_slot_count;
    END IF;
END $$;

ALTER TABLE reminders
    ADD CONSTRAINT reminders_task_user_type_remind_at_key
        UNIQUE (task_id, user_id, type, remind_at);
