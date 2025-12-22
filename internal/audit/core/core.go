package core

import (
	"github.com/C5rogers/G-Synch/internal/models"
)

type Audit interface {
	Check(target, given Schema) ([]models.CheckReturn, error)
	ReverseCheck(target, given Schema) ([]string, error)
	Sync(targetAdapter SchemaAdapter, given SchemaAdapter) ([]string, error)
	Name() string
}

// this interface will be applied for different adapters like Postgres, MySQL, etc
