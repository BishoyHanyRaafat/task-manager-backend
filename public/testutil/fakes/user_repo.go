package fakes

import (
	"context"
	"errors"
	"sync"
	"task_manager/public/repositories"
	"task_manager/public/repositories/models"

	"github.com/google/uuid"
)

var _ repositories.UserRepository = (*UserRepo)(nil)

type UserRepo struct {
	mu      sync.Mutex
	byEmail map[string]*models.User
	byID    map[uuid.UUID]*models.User
	pw      map[uuid.UUID]string
	byProv  map[string]uuid.UUID                                  // key: string(provider) + ":" + provider_user_id
	provs   map[uuid.UUID]map[models.Provider]models.AuthProvider // user_id -> provider -> provider
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byEmail: make(map[string]*models.User),
		byID:    make(map[uuid.UUID]*models.User),
		pw:      make(map[uuid.UUID]string),
		byProv:  make(map[string]uuid.UUID),
		provs:   make(map[uuid.UUID]map[models.Provider]models.AuthProvider),
	}
}

func (r *UserRepo) CreateUser(_ context.Context, u *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *u
	r.byEmail[u.Email] = &clone
	r.byID[u.ID] = &clone
	return nil
}

func (r *UserRepo) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u := r.byEmail[email]
	if u == nil {
		return nil, nil
	}
	clone := *u
	return &clone, nil
}

func (r *UserRepo) GetUserByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u := r.byID[id]
	if u == nil {
		return nil, nil
	}
	clone := *u
	return &clone, nil
}

func (r *UserRepo) UpsertPassword(_ context.Context, userID uuid.UUID, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pw[userID] = passwordHash
	return nil
}

func (r *UserRepo) GetPasswordHashByUserID(_ context.Context, userID uuid.UUID) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pw[userID], nil
}

func (r *UserRepo) GetUserByAuthProvider(_ context.Context, provider models.Provider, providerUserID string) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := string(provider) + ":" + providerUserID
	uid, ok := r.byProv[key]
	if !ok {
		return nil, nil
	}
	u := r.byID[uid]
	if u == nil {
		return nil, nil
	}
	clone := *u
	return &clone, nil
}

func (r *UserRepo) GetAuthProviderByUserAndProvider(_ context.Context, userID uuid.UUID, provider models.Provider) (*models.AuthProvider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := r.provs[userID]
	if m == nil {
		return nil, nil
	}
	ap, ok := m[provider]
	if !ok {
		return nil, nil
	}
	clone := ap
	return &clone, nil
}

func (r *UserRepo) ListAuthProvidersByUserID(_ context.Context, userID uuid.UUID) ([]models.AuthProvider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := r.provs[userID]
	if m == nil {
		return []models.AuthProvider{}, nil
	}
	out := make([]models.AuthProvider, 0, len(m))
	for _, ap := range m {
		out = append(out, ap)
	}
	return out, nil
}

func (r *UserRepo) CreateAuthProvider(_ context.Context, ap *models.AuthProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := string(ap.Provider) + ":" + ap.ProviderUserID
	if existingUID, ok := r.byProv[key]; ok && existingUID != ap.UserID {
		return errors.New("unique violation: provider_user_id already linked")
	}
	if r.provs[ap.UserID] == nil {
		r.provs[ap.UserID] = make(map[models.Provider]models.AuthProvider)
	}
	if existing, ok := r.provs[ap.UserID][ap.Provider]; ok && existing.ProviderUserID != ap.ProviderUserID {
		return errors.New("unique violation: user already has provider linked")
	}
	clone := *ap
	r.provs[ap.UserID][ap.Provider] = clone
	r.byProv[key] = ap.UserID
	return nil
}
