CREATE TABLE team_members (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    role ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_team_members_unique (user_id, team_id),
    CONSTRAINT fk_team_members_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_team_members_team FOREIGN KEY (team_id) REFERENCES teams(id)
);
