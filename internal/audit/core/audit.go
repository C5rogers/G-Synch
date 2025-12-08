package core

type Audit interface {
	Check(target, given Schema) ([]string, error)
	ReverseCheck(target, given Schema) ([]string, error)
	Sync(targetAdapter SchemaAdapter, given SchemaAdapter) ([]string, error)
	Name() string
}
