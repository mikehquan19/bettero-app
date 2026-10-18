package services

import (
	"betterov2/models"
	"betterov2/repositories"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SessionDuration is the amount of time a login session remains valid.
const SessionDuration = 30 * 24 * time.Hour

// AuthService handles user registration, login, session authentication, and logout.
type AuthService struct {
	db   *pgxpool.Pool
	repo *repositories.AuthRepo
}

// NewAuthService creates an AuthService with its database connection and repository.
func NewAuthService(db *pgxpool.Pool, repo *repositories.AuthRepo) *AuthService {
	return &AuthService{
		db:   db,
		repo: repo,
	}
}

// Register validates the registration details, creates the user, and starts a session.
// It returns the new user and the session token.
func (s *AuthService) Register(ctx context.Context, body models.RegisterUserBody) (models.User, string, error) {
	body.FirstName = strings.TrimSpace(body.FirstName)

	// First and last names are required and must fit the database columns.
	if body.FirstName == "" || len(body.FirstName) > 50 {
		return models.User{}, "", models.ErrInvalidUserBody
	}
	body.LastName = strings.TrimSpace(body.LastName)
	if body.LastName == "" || len(body.LastName) > 50 {
		return models.User{}, "", models.ErrInvalidUserBody
	}

	// Usernames must be between 3 and 50 characters.
	body.Username = strings.ToLower(strings.TrimSpace(body.Username))
	if len(body.Username) < 3 || len(body.Username) > 50 {
		return models.User{}, "", models.ErrInvalidUserBody
	}

	// Email is optional, but must fit the database column when provided.
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	if len(body.Email) > 100 {
		return models.User{}, "", models.ErrInvalidUserBody
	}

	// bcrypt only supports passwords up to 72 bytes.
	if len(body.Password) < 10 || len(body.Password) > 72 {
		return models.User{}, "", models.ErrInvalidUserBody
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, "", err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, "", err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Create the user with the hashed password.
	user, err := s.repo.CreateUser(ctx, tx, body, string(passwordHash))
	if err != nil {
		return models.User{}, "", err
	}

	// Generate a token for the user's login session.
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return models.User{}, "", err
	}
	expiresAt := time.Now().Add(SessionDuration)
	if err = s.repo.CreateSession(ctx, tx, tokenHash, user.ID, expiresAt); err != nil {
		return models.User{}, "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return models.User{}, "", err
	}

	return user, token, nil
}

// Login checks the user's password and starts a new session.
// It returns the authenticated user and the session token.
func (s *AuthService) Login(ctx context.Context, body models.LoginBody) (models.User, string, error) {
	username := strings.ToLower(strings.TrimSpace(body.Username))
	user, err := s.repo.GetUserByUsername(ctx, s.db, username)
	if err != nil {
		return models.User{}, "", err
	}

	// Compare the supplied password with the stored password hash.
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return models.User{}, "", models.ErrInvalidCredentials
	}

	// Start a new session after successful authentication.
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return models.User{}, "", err
	}
	expiresAt := time.Now().Add(SessionDuration)
	if err = s.repo.CreateSession(ctx, s.db, tokenHash, user.ID, expiresAt); err != nil {
		return models.User{}, "", err
	}

	return user, token, nil
}

// Authenticate looks up the user associated with a session token.
// It returns ErrNotFound if the token is empty, expired, or unknown.
func (s *AuthService) Authenticate(ctx context.Context, token string) (models.User, error) {
	if token == "" {
		return models.User{}, models.ErrNotFound
	}

	hash := sha256.Sum256([]byte(token))
	return s.repo.UserForSession(ctx, s.db, hash[:])
}

// Logout deletes the session associated with the token.
// An empty token is treated as an already-ended session.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	hash := sha256.Sum256([]byte(token))
	return s.repo.DeleteSession(ctx, s.db, hash[:])
}

// newSessionToken creates a cryptographically random session token and its hash.
// The raw token is returned to the caller; only the hash should be stored.
func newSessionToken() (string, []byte, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", nil, err
	}

	token := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	return token, hash[:], nil
}
