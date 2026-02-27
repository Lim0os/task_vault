package mysql

import (
	"context"
	"database/sql"
	"task_vault/internal/domain"

	"github.com/google/uuid"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, team *domain.Team) error {
	team.ID = uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO teams (id, name, created_by) VALUES (?, ?, ?)",
		team.ID, team.Name, team.CreatedBy,
	)
	return err
}

func (r *TeamRepo) AddMember(ctx context.Context, member *domain.TeamMember) error {
	member.ID = uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO team_members (id, user_id, team_id, role) VALUES (?, ?, ?, ?)",
		member.ID, member.UserID, member.TeamID, member.Role,
	)
	return err
}

func (r *TeamRepo) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	team := &domain.Team{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_by, created_at FROM teams WHERE id = ?", id,
	).Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepo) ListByUser(ctx context.Context, userID string) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.name, t.created_by, t.created_at
		 FROM teams t
		 JOIN team_members tm ON tm.team_id = t.id
		 WHERE tm.user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepo) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	m := &domain.TeamMember{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, team_id, role, joined_at FROM team_members WHERE team_id = ? AND user_id = ?",
		teamID, userID,
	).Scan(&m.ID, &m.UserID, &m.TeamID, &m.Role, &m.JoinedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}
