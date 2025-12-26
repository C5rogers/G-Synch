package core

import (
	"github.com/C5rogers/G-Synch/internal/models"
)

type Audit interface {
	/*
	 * Check checks the differences between the target and given schemas.
	 */
	Check(target, given Schema) ([]models.CheckReturn, error)
	/*
	 * ReverseCheck checks the differences between the given and target schemas.
	 */
	ReverseCheck(target, given Schema) ([]string, error)
	/*
	 * Sync synchronizes the target and given schemas.
	 */
	Sync(targetAdapter SchemaAdapter, given SchemaAdapter) ([]string, error)
	/*
	 * Name returns the name of the audit.
	 */
	Name() string
}
