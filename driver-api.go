package sqldriver

import (
	"database/sql"
)

type ExecResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type ResultSet interface {
	Columns() []string
	Close()   error
	Next(dest []interface{}) error
}

type DriverWrapper interface {
	GetDriverName() string
	CreateConnection(dsn string) (interface{}, error)
	CloseConnection(conn interface{}) error
	Ping(conn interface{}) error
	BeginTx(conn interface{}) (interface{}, error)
	Commit(tx interface{}) error
	Rollback(tx interface{}) error
	Exec(conn interface{}, sql string, args ...interface{}) (ExecResult, error)
	Query(conn interface{}, sql string, args ...interface{}) (ResultSet, error)
}

func Register(wrapper DriverWrapper) {
	sql.Register(wrapper.GetDriverName(), &innerDriver{wrapper})
}
