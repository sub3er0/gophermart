package interfaces

import (
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DBStorageInterface interface {
	Init(connectionString string) error
	Exec(query string, args ...interface{}) (pgconn.CommandTag, error)
	Select(query string, args ...interface{}) (pgx.Rows, error)
	Close()
	BeginTransaction() error
	Rollback() error
	Commit() error
}
