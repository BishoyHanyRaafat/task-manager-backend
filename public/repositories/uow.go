package repositories

import (
	"context"
	"database/sql"
	"fmt"

	dbx "task_manager/public/db"
)

// Repos groups all repositories that should share the same DB handle (DB or Tx).
// Over time you can add more repos here (Teams, Tasks, etc).
type Repos struct {
	Users UserRepository
	Teams TeamRepository
}

// Transaction is an explicit, manually-managed transaction scope.
//
// Typical usage:
//
//	tx, err := uow.Begin(ctx)
//	if err != nil { ... }
//	defer tx.Stop() // rollback unless committed
//
//	if err := tx.Users().CreateUser(ctx, u); err != nil { ... }
//	// ... any logic in-between ...
//	if err := tx.Users().UpsertPassword(ctx, u.ID, hash); err != nil { ... }
//
//	if err := tx.Commit(); err != nil { ... }
type Transaction interface {
	Repos() Repos
	Users() UserRepository
	Teams() TeamRepository
	Commit() error
	Rollback() error
	Stop() error
}

// UnitOfWork provides non-transactional repos by default, plus an explicit transaction wrapper.
//
// Normal usage (NO transaction):
//
//	users := uow.Users()
//
// Explicit transaction usage:
//
//	err := uow.WithTransaction(ctx, func(ctx context.Context, r repositories.Repos) error {
//		if err := r.Users.CreateUser(ctx, u); err != nil {
//			return err
//		}
//		// ... do any logic in-between ...
//		return r.Users.CreateAuthProvider(ctx, ap)
//	})
type UnitOfWork interface {
	Users() UserRepository
	Teams() TeamRepository
	Begin(ctx context.Context) (Transaction, error)
	WithTransaction(ctx context.Context, fn func(ctx context.Context, r Repos) error) error
}

type unitOfWork struct {
	driver string
	db     *sql.DB
}

type transaction struct {
	tx        *sql.Tx
	repos     Repos
	committed bool
}

func NewUnitOfWork(driver string, db *sql.DB) (UnitOfWork, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	// Validate driver early so Users() can be error-free.
	if _, err := NewUserRepositoryWithDBTX(driver, db); err != nil {
		return nil, err
	}
	return &unitOfWork{driver: driver, db: db}, nil
}

func (u *unitOfWork) Users() UserRepository {
	// Driver is validated in NewUnitOfWork; ignore error here.
	repos, _ := u.buildRepos(u.db)
	return repos.Users
}

func (u *unitOfWork) Teams() TeamRepository {
	// Driver is validated in NewUnitOfWork; ignore error here.
	repos, _ := u.buildRepos(u.db)
	return repos.Teams
}

func (u *unitOfWork) Begin(ctx context.Context) (Transaction, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	repos, err := u.buildRepos(tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return &transaction{tx: tx, repos: repos}, nil
}

// buildRepos creates all repositories with the given DB handle.
// Add new repositories here when extending the system.
func (u *unitOfWork) buildRepos(db dbx.DBTX) (Repos, error) {
	users, err := NewUserRepositoryWithDBTX(u.driver, db)
	if err != nil {
		return Repos{}, err
	}
	teams, err := NewTeamRepositoryWithDBTX(u.driver, db)
	if err != nil {
		return Repos{}, err
	}
	return Repos{Users: users, Teams: teams}, nil
}

func (u *unitOfWork) WithTransaction(ctx context.Context, fn func(ctx context.Context, r Repos) error) error {
	if fn == nil {
		return fmt.Errorf("transaction fn is nil")
	}

	txn, err := u.Begin(ctx)
	if err != nil {
		return err
	}
	defer txn.Stop()

	if err := fn(ctx, txn.Repos()); err != nil {
		return err
	}
	return txn.Commit()
}

func (t *transaction) Repos() Repos {
	return t.repos
}

func (t *transaction) Users() UserRepository {
	return t.repos.Users
}

func (t *transaction) Teams() TeamRepository {
	return t.repos.Teams
}

func (t *transaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if t.committed {
		return nil
	}
	if err := t.tx.Commit(); err != nil {
		return err
	}
	t.committed = true
	return nil
}

func (t *transaction) Rollback() error {
	if t.tx == nil || t.committed {
		return nil
	}
	err := t.tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		return err
	}
	return nil
}

func (t *transaction) Stop() error {
	// By convention: rollback unless committed.
	return t.Rollback()
}
