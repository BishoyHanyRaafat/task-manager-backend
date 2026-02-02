package fakes

import (
	"context"
	"fmt"
	"task_manager/public/repositories"
)

var _ repositories.UnitOfWork = (*UnitOfWork)(nil)

type transaction struct {
	users     repositories.UserRepository
	committed bool
}

var _ repositories.Transaction = (*transaction)(nil)

// UnitOfWork is a lightweight test double that executes "transaction" ops
// against the same in-memory repositories without actually opening a DB transaction.
type UnitOfWork struct {
	users repositories.UserRepository
}

func NewUnitOfWork(users repositories.UserRepository) *UnitOfWork {
	return &UnitOfWork{users: users}
}

func (u *UnitOfWork) Users() repositories.UserRepository {
	return u.users
}

func (u *UnitOfWork) Begin(_ context.Context) (repositories.Transaction, error) {
	return &transaction{users: u.users}, nil
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
	return repositories.Repos{Users: t.users}
}

func (t *transaction) Users() repositories.UserRepository {
	return t.users
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
