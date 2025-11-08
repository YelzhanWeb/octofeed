package ports

import "time"

type ConfigProvider interface {
	GetPostgresHost() string
	GetPostgresPort() string
	GetPostgresUser() string
	GetPostgresPassword() string
	GetPostgresDBName() string
	GetPostgresSSLMode() string
	GetDefaultInterval() time.Duration
	GetDefaultWorkersCount() int
}
