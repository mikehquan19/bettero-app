package repositories

import (
	"betterov2/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthRepo struct{}

func NewAuthRepo() *AuthRepo {
	return &AuthRepo{}
}

func (r *AuthRepo) CreateUser(
	ctx context.Context,
	dbtx DBTX,
	body models.RegisterUserBody,
	passwordHash string,
) (models.User, error) {
	var user models.User

	var email any
	if body.Email != "" {
		email = body.Email
	}

	const createUserSQL = `
		INSERT INTO users (
			first_name,
			last_name,
			username,
			email,
			password_hash
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			first_name,
			last_name,
			username,
			COALESCE(email, ''),
			password_hash,
			created_at,
			updated_at;
	`

	row := dbtx.QueryRow(
		ctx,
		createUserSQL,
		body.FirstName,
		body.LastName,
		body.Username,
		email,
		passwordHash,
	)

	if err := scanUser(row, &user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, models.ErrUserExists
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *AuthRepo) GetUserByUsername(
	ctx context.Context,
	dbtx DBTX,
	username string,
) (models.User, error) {
	var user models.User

	const getUserByUsernameSQL = `
		SELECT
			id,
			first_name,
			last_name,
			username,
			COALESCE(email, ''),
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE username = $1;
	`

	row := dbtx.QueryRow(ctx, getUserByUsernameSQL, username)
	if err := scanUser(row, &user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrInvalidCredentials
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *AuthRepo) CreateSession(
	ctx context.Context,
	dbtx DBTX,
	tokenHash []byte,
	userID uuid.UUID,
	expiresAt time.Time,
) error {
	const deleteExpiredSessionsSQL = "DELETE FROM user_sessions WHERE expires_at <= NOW();"
	if _, err := dbtx.Exec(ctx, deleteExpiredSessionsSQL); err != nil {
		return err
	}

	const createSessionSQL = `
		INSERT INTO user_sessions (
			token_hash,
			user_id,
			expires_at
		)
		VALUES ($1, $2, $3);
	`
	_, err := dbtx.Exec(ctx, createSessionSQL, tokenHash, userID, expiresAt)
	return err
}

func (r *AuthRepo) UserForSession(
	ctx context.Context,
	dbtx DBTX,
	tokenHash []byte,
) (models.User, error) {
	var user models.User

	const userForSessionSQL = `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			COALESCE(u.email, ''),
			u.password_hash,
			u.created_at,
			u.updated_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW();
	`

	row := dbtx.QueryRow(ctx, userForSessionSQL, tokenHash)
	if err := scanUser(row, &user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *AuthRepo) DeleteSession(ctx context.Context, dbtx DBTX, tokenHash []byte) error {
	const deleteSessionSQL = "DELETE FROM user_sessions WHERE token_hash = $1;"
	_, err := dbtx.Exec(ctx, deleteSessionSQL, tokenHash)
	return err
}

func scanUser(userRow pgx.Row, user *models.User) error {
	return userRow.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}
