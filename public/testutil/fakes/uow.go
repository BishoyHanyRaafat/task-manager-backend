package fakes

import (
	"context"
	"fmt"
	"task_manager/public/repositories"
)

var _ repositories.UnitOfWork = (*UnitOfWork)(nil)

type transaction struct {
	users     repositories.UserRepository
	teams     repositories.TeamRepository
	committed bool
}

var _ repositories.Transaction = (*transaction)(nil)

// UnitOfWork is a lightweight test double that executes "transaction" ops
// against the same in-memory repositories without actually opening a DB transaction.
type UnitOfWork struct {
	users repositories.UserRepository
	teams repositories.TeamRepository
}

func NewUnitOfWork(users repositories.UserRepository, teams repositories.TeamRepository) *UnitOfWork {
	return &UnitOfWork{users: users, teams: teams}
}

func (u *UnitOfWork) Users() repositories.UserRepository {
	return u.users
}

func (u *UnitOfWork) Teams() repositories.TeamRepository {
	return u.teams
}

func (u *UnitOfWork) Begin(_ context.Context) (repositories.Transaction, error) {
	return &transaction{users: u.users, teams: u.teams}, nil
}

func (u *UnitOfWork) WithTransaction(ctx context.Context, fn func(ctx context.Context, r repositories.Repos) error) error {
	if fn == nil {
		return fmt.Errorf("transaction fn is nil")
	}
	tx, _ := u.Begin(ctx)
	defer tx.Stop()
	if err := fn(ctx, tx.Repos()); err != nil {
		return err
	}
	return tx.Commit()
}

func (t *transaction) Repos() repositories.Repos {
	return repositories.Repos{Users: t.users, Teams: t.teams}
}

func (t *transaction) Users() repositories.UserRepository {
	return t.users
}

func (t *transaction) Teams() repositories.TeamRepository {
	return t.teams
}

func (t *transaction) Commit() error {
	t.committed = true
	return nil
}

func (t *transaction) Rollback() error {
	// no-op; in-memory repo
	return nil
}

func (t *transaction) Stop() error {
	if t.committed {
		return nil
	}
	return t.Rollback()
}
