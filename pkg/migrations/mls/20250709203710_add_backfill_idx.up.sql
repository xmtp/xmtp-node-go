CREATE INDEX idx_group_messages_is_commit_null_id_asc
    ON group_messages (id)
    WHERE is_commit IS NULL;