package postgress

import (
	"context"
	"database/sql"
	"errors"
	"task_manager/public/repositories/models"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, first_name, last_name, email, user_type) VALUES ($1, $2, $3, $4, $5)`,
		u.ID,
		u.FirstName,
		u.LastName,
		u.Email,
		u.UserType,
	)
	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, first_name, last_name, email, created_at, updated_at, user_type FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.UserType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, first_name, last_name, email, created_at, updated_at, user_type FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.UserType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpsertPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	now := time.Now()
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO passwords (id, user_id, v, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $4)
		 ON CONFLICT (user_id) DO UPDATE
		 SET v = EXCLUDED.v, updated_at = EXCLUDED.updated_at`,
		uuid.New(),
		userID,
		passwordHash,
		now,
	)
	return err
}

func (r *UserRepository) GetPasswordHashByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT v FROM passwords WHERE user_id = $1`, userID).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return v, nil
}

func (r *UserRepository) GetUserByAuthProvider(ctx context.Context, provider models.Provider, providerUserID string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.first_name, u.last_name, u.email, u.created_at, u.updated_at, u.user_type
		 FROM auth_providers ap
		 JOIN users u ON u.id = ap.user_id
		 WHERE ap.provider = $1 AND ap.provider_user_id = $2`,
		provider,
		providerUserID,
	).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.UserType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetAuthProviderByUserAndProvider(ctx context.Context, userID uuid.UUID, provider models.Provider) (*models.AuthProvider, error) {
	var ap models.AuthProvider
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at, updated_at
		 FROM auth_providers
		 WHERE user_id = $1 AND provider = $2`,
		userID,
		provider,
	).Scan(&ap.ID, &ap.UserID, &ap.Provider, &ap.ProviderUserID, &ap.Email, &ap.Username, &ap.DisplayName, &ap.AvatarURL, &ap.CreatedAt, &ap.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ap, nil
}

func (r *UserRepository) ListAuthProvidersByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthProvider, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, provider, provider_user_id, COALESCE(email, ''), COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at, updated_at
		 FROM auth_providers
		 WHERE user_id = $1
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AuthProvider
	for rows.Next() {
		var ap models.AuthProvider
		if err := rows.Scan(&ap.ID, &ap.UserID, &ap.Provider, &ap.ProviderUserID, &ap.Email, &ap.Username, &ap.DisplayName, &ap.AvatarURL, &ap.CreatedAt, &ap.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *UserRepository) CreateAuthProvider(ctx context.Context, ap *models.AuthProvider) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO auth_providers (id, user_id, provider, provider_user_id, email, username, display_name, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9, $10)`,
		ap.ID,
		ap.UserID,
		ap.Provider,
		ap.ProviderUserID,
		ap.Email,
		ap.Username,
		ap.DisplayName,
		ap.AvatarURL,
		ap.CreatedAt,
		ap.UpdatedAt,
	)
	return err
}
