package sqldriver

import (
	"database/sql/driver"
	"runtime"
	"context"
)

// -------------- inner wrapper ------------
type innerConn struct {
	conn    interface{}
	wrapper DriverWrapper
}

type innerDriver struct {
	wrapper DriverWrapper
}

type innerStmt struct {
	c *innerConn
	stmt interface{}
	q string
}

type innerTx struct {
	tx interface{}
	c *innerConn
}

type innerRows struct {
	rs ResultSet
}

// implementation of database/sql/driver.Open
func (inD *innerDriver) Open(dsn string) (driver.Conn, error) {
	if conn, err := inD.wrapper.CreateConnection(dsn); err != nil {
		return nil, err
	} else {
		c := &innerConn{conn, inD.wrapper}
		runtime.SetFinalizer(c, (*innerConn).Close)
		return c, nil
	}
}

// implementation of database/sql/driver.Conn
func (inC *innerConn) Prepare(query string) (driver.Stmt, error) {
	if s, err := inC.wrapper.Prepare(inC.conn, query); err != nil {
		return nil, err
	} else {
		stmt := &innerStmt{inC, s, query}
		runtime.SetFinalizer(stmt, (*innerStmt).Close)
		return stmt, nil
	}
}

func (inC *innerConn) Close() error {
	err := inC.wrapper.CloseConnection(inC.conn)
	// runtime.SetFinalizer(inC, nil)
	return err
}

func (inC *innerConn) Begin() (driver.Tx, error) {
	if tx, err := inC.wrapper.BeginTx(inC.conn); err != nil {
		return nil, err
	} else {
		return &innerTx{tx, inC}, nil
	}
}

func (inC *innerConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return inC.Begin()
}

// implementation of dabase/sql/driver.Pinger
func (inC *innerConn) Ping(ctx context.Context) (err error) {
	return inC.wrapper.Ping(inC.conn)
}

// implementation of database/sql/driver.Stmt
func (inStmt *innerStmt) Close() error {
	return inStmt.c.wrapper.CloseStmt(inStmt.stmt)
}

func (inStmt *innerStmt) NumInput() int {
	return -1
}

func (inStmt *innerStmt) Exec(args []driver.Value) (driver.Result, error) {
	var as []interface{}
	if args != nil {
		as = make([]interface{}, len(args))
		for i, arg := range args {
			as[i] = arg
		}
	}
	r, e := inStmt.c.wrapper.Exec(inStmt.stmt, as...)
	return r, e
}

func (inStmt *innerStmt) Query(args []driver.Value) (driver.Rows, error) {
	var as []interface{}
	if args != nil {
		as = make([]interface{}, len(args))
		for i, arg := range args {
			as[i] = arg
		}
	}
	r, e := inStmt.c.wrapper.Query(inStmt.stmt, as...)
	return wrapperRows(r), e
}

// implementation of database/sql/driver.Tx
func (inTx *innerTx) Commit() error {
	return inTx.c.wrapper.Commit(inTx.tx)
}

func (inTx *innerTx) Rollback() error {
	return inTx.c.wrapper.Rollback(inTx.tx)
}

// implementation of database/sql/driver.Rows
func wrapperRows(rows ResultSet) driver.Rows {
	return &innerRows{rows}
}

func (r *innerRows) Columns() []string {
	return r.rs.Columns()
}

func (r *innerRows) Close() error {
	return r.rs.Close()
}

func (r *innerRows) Next(dest []driver.Value) error {
	if dest == nil {
		return r.rs.Next(nil)
	}
	args := make([]interface{}, len(dest))
	if err := r.rs.Next(args); err != nil {
		return err
	}
	for i, arg := range args {
		dest[i] = arg
	}
	return nil
}
