ALTER TABLE task_comments DROP FOREIGN KEY fk_task_comments_task;
ALTER TABLE task_comments DROP FOREIGN KEY fk_task_comments_user;

ALTER TABLE task_history DROP FOREIGN KEY fk_task_history_task;
ALTER TABLE task_history DROP FOREIGN KEY fk_task_history_user;

ALTER TABLE tasks DROP FOREIGN KEY fk_tasks_assignee;
ALTER TABLE tasks DROP FOREIGN KEY fk_tasks_team;
ALTER TABLE tasks DROP FOREIGN KEY fk_tasks_created_by;

ALTER TABLE team_members DROP FOREIGN KEY fk_team_members_user;
ALTER TABLE team_members DROP FOREIGN KEY fk_team_members_team;

ALTER TABLE teams DROP FOREIGN KEY fk_teams_created_by;

ALTER TABLE users MODIFY id CHAR(36) NOT NULL;

ALTER TABLE teams MODIFY id CHAR(36) NOT NULL;
ALTER TABLE teams MODIFY created_by CHAR(36) NOT NULL;

ALTER TABLE team_members MODIFY id CHAR(36) NOT NULL;
ALTER TABLE team_members MODIFY user_id CHAR(36) NOT NULL;
ALTER TABLE team_members MODIFY team_id CHAR(36) NOT NULL;

ALTER TABLE tasks MODIFY id CHAR(36) NOT NULL;
ALTER TABLE tasks MODIFY assignee_id CHAR(36);
ALTER TABLE tasks MODIFY team_id CHAR(36) NOT NULL;
ALTER TABLE tasks MODIFY created_by CHAR(36) NOT NULL;

ALTER TABLE task_history MODIFY id CHAR(36) NOT NULL;
ALTER TABLE task_history MODIFY task_id CHAR(36) NOT NULL;
ALTER TABLE task_history MODIFY changed_by CHAR(36) NOT NULL;

ALTER TABLE task_comments MODIFY id CHAR(36) NOT NULL;
ALTER TABLE task_comments MODIFY task_id CHAR(36) NOT NULL;
ALTER TABLE task_comments MODIFY user_id CHAR(36) NOT NULL;

ALTER TABLE teams ADD CONSTRAINT fk_teams_created_by FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE team_members ADD CONSTRAINT fk_team_members_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_team FOREIGN KEY (team_id) REFERENCES teams(id);

ALTER TABLE tasks ADD CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES users(id);
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_team FOREIGN KEY (team_id) REFERENCES teams(id);
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_created_by FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE task_history ADD CONSTRAINT fk_task_history_task FOREIGN KEY (task_id) REFERENCES tasks(id);
ALTER TABLE task_history ADD CONSTRAINT fk_task_history_user FOREIGN KEY (changed_by) REFERENCES users(id);

ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_task FOREIGN KEY (task_id) REFERENCES tasks(id);
ALTER TABLE task_comments ADD CONSTRAINT fk_task_comments_user FOREIGN KEY (user_id) REFERENCES users(id);
