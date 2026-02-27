package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"task_vault/internal/domain"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) conn(ctx context.Context) DBTX {
	return ConnFromContext(ctx, r.db)
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	user.ID = uuid.New().String()
	_, err := r.conn(ctx).ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, name) VALUES (?, ?, ?, ?)",
		user.ID, user.Email, user.PasswordHash, user.Name,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	err := r.conn(ctx).QueryRowContext(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE id = ?", id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("получение пользователя [id=%s]: %w", id, err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.conn(ctx).QueryRowContext(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE email = ?", email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("получение пользователя [email=%s]: %w", email, err)
	}
	return user, nil
}
