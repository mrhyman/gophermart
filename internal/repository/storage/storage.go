package storage

type Storage interface {
	HealthChecker
}

type HealthChecker interface {
	Ping() error
	Close() error
}
