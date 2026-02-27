CREATE TABLE task_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    changed_by BIGINT NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_history_task_changed (task_id, changed_at),
    CONSTRAINT fk_task_history_task FOREIGN KEY (task_id) REFERENCES tasks(id),
    CONSTRAINT fk_task_history_user FOREIGN KEY (changed_by) REFERENCES users(id)
);
