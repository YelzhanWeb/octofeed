package ports

type Migrator interface {
	Up() error
	Down() error
}
