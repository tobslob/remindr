ALTER TABLE tags
ADD CONSTRAINT tags_user_id_name_key UNIQUE (user_id, name);
