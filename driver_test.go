package sqldriver

import (
	"database/sql/driver"
	"database/sql"
	"fmt"
	"io"
	"testing"
	//"github.com/go-xorm/xorm"
	//"github.com/go-xorm/core"
)

// --------  a test driver implementation  -----------
const (
	DRIVER_NAME = "sqltest"
)

type TResult struct {
}

func (er *TResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (er * TResult) RowsAffected() (int64, error) {
	return 0, nil
}

type TResultSet struct {
	count int
}

func (rs *TResultSet) Columns() []string {
	return []string{"id", "name", "age"}
}

func (rs *TResultSet) Close() error {
	return nil
}

func (rs *TResultSet) Next(dest []interface{}) error {
	if (rs.count > 0) {
		dest[0] = rs.count
		dest[1] = fmt.Sprintf("hello %d", rs.count)
		dest[2] = 10
		rs.count--
		return nil
	}
	return io.EOF
}

type TDriver struct {
	connected bool
	count int
}

func (d *TDriver) GetDriverName() string {
	return DRIVER_NAME
}

func (d *TDriver) CreateConnection(dsn string) (interface{}, error) {
	fmt.Printf("CreateConnection(%s) called\n", dsn)
	if !d.connected {
		fmt.Printf("create new connection\n")
		d.connected = true
		d.count = 1
	} else {
		fmt.Printf("using exiting connection\n")
		d.count++
	}
	return true, nil
}

func (d *TDriver) CloseConnection(conn interface{}) error {
	fmt.Printf("CloseConnection called\n")
	if d.count > 0 {
		fmt.Printf("count: %d, dec it\n", d.count)
		d.count--
		if d.count == 0 {
			fmt.Printf("real close\n")
			d.connected = false
		}
		return nil
	}
	fmt.Printf("no close any more\n")
	return fmt.Errorf("don't call too more closing\n")
}

func (d *TDriver) Ping(conn interface{}) error {
	return nil
}

func (d *TDriver) BeginTx(conn interface{}, opts driver.TxOptions) (interface{}, error) {
	fmt.Printf("BeginTx called\n")
	return nil, nil
}

func (d *TDriver) Commit(tx interface{}) error {
	fmt.Printf("Commit called\n")
	return nil
}

func (d *TDriver) Rollback(tx interface{}) error {
	fmt.Printf("Rollback called\n")
	return nil
}

func (d *TDriver) Prepare(conn interface{}, query string) (interface{}, error) {
	fmt.Printf("Prepare %s called\n", query)
	return nil, nil
}

func (d *TDriver) CloseStmt(stmt interface{}) error {
	fmt.Printf("CloseStmt called\n")
	return nil
}

func (d *TDriver) Exec(stmt interface{}, args ...interface{}) (ExecResult, error) {
	fmt.Printf("exec called\n")
	return &TResult{}, nil
}

func (d *TDriver) Query(stmt interface{}, args ...interface{}) (ResultSet, error) {
	fmt.Printf("query called\n")
	return &TResultSet{3}, nil
}

/*
// Parse() defined in xorm.Driver
func (d *TDriver) Parse(_ string, _ string) (*core.Uri, error) {
	return &core.Uri{DbType:"mysql"}, nil
}*/

// --------------- testing ------------------
var _testDriver = &TDriver{}
func Test_driverRunning(t *testing.T) {
	Register(_testDriver) // in real driver, call Register() in func init()
	conn, err := sql.Open(DRIVER_NAME, "this-is-a-test-driver")
	if err != nil {
		t.Fatalf("failed to open: %v\n", err)
	}
	defer conn.Close()

	/*
	stmt, err := conn.Prepare("test query")
	if err != nil {
		t.Fatalf("failed to prepare: %v\n", err)
	}
	defer stmt.Close()
	rows, err := stmt.Query("any-args-is-ok", 1, true)
	*/
	rows, err := conn.Query("test query", "any-args-is-ok", 1, true)
	if err != nil {
		t.Fatalf("failed to query: %v\n", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("failed to get columns: %v\n", err)
	}
	fmt.Printf("columns: %v\n", cols)

	colNum := len(cols)
	scanArgs := make([]interface{}, colNum)
	row := make([]interface{}, colNum)
	for i := range row {
		scanArgs[i] = &row[i]
	}

	for rows.Next() {
		rows.Scan(scanArgs...)
		for i, col := range cols {
			fmt.Printf("[%s]: %v\t", col, row[i])
		}
		fmt.Printf("\n")
	}
}

/*
func Test_xormRunning(t *testing.T) {
	// Register(&TDriver{}) // in real driver, call Register() in func init()
	core.RegisterDriver(DRIVER_NAME, _testDriver)
	db, err := xorm.NewEngine(DRIVER_NAME, "this-is-a-test-driver")
	if err != nil {
		t.Fatalf("failed to open: %v\n", err)
	}
	db.ShowSQL(true)
	var rows []map[string]interface{}
	if err = db.Table("table").Find(&rows); err != nil {
		t.Fatalf("failed to get rows: %v\n", err)
	}
	fmt.Printf("%v\n", rows)

	session := db.NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		// if returned then will rollback automatically
		t.Fatalf("failed to Begin(): %v\n", err)
	}

	var user = struct {
		Id int64
		Name string
	} {
		Id: int64(3),
		Name: "hello",
	}
	result, err := session.Table("hello").Update(&user)
    if err != nil {
        t.Fatalf("%v\n", err)
    }
    fmt.Println("result:", result)
	session.Commit()
}*/
