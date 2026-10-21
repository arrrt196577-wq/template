// Package transaction provides an explicit transaction boundary for future services.
package transaction

import (
	"context"

	"gorm.io/gorm"
)

type Runner struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

// Within passes the only database handle repositories should use for this operation.
// Bind repositories to tx inside the callback; never fall back to the original db.
func (runner *Runner) Within(ctx context.Context, operation func(tx *gorm.DB) error) error {
	return runner.db.WithContext(ctx).Transaction(operation)
}
