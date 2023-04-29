package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"log"
	"sqlow/helpers"
	"strings"
)

type Driver struct {
	Engine      string
	Host        string
	Port        string
	Schema      string
	Username    string
	Password    string
	Options     []string
	IsDryRun    bool
	connection  *sql.DB
	transaction *sql.Tx
}

type Conn interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

func (d *Driver) Connect() {
	d.transaction = nil
	switch d.Engine {
	case "postgres":
		d.connection = getPostgresConnection(*d)
	case "maria", "mariadb", "mysql":
		d.connection = getMariaConnection(*d)
	default:
		log.Fatalf("Unsupported engine type: %s", d.Engine)
	}
	ok, hasRows := d.QueryPasses("select 1;")
	if !ok || !hasRows {
		log.Fatalf("Unable to connect to %s database", d.Engine)
	}
	if d.IsDryRun {
		log.Println("Beginning transaction...")
		d.BeginTransaction()
	}
}

func (d *Driver) BeginTransaction() {
	tx, err := d.connection.BeginTx(context.Background(), nil)
	helpers.CheckError(err)
	d.transaction = tx
}

func (d *Driver) RollbackTransaction() {
	err := d.transaction.Rollback()
	helpers.CheckWarn(err)
}

func (d *Driver) Savepoint() {
	_, err := d.conn().Exec("SAVEPOINT sqlow_save;")
	helpers.CheckError(err)
}

func (d *Driver) Loadpoint() {
	_, err := d.conn().Exec("ROLLBACK TO sqlow_save;")
	helpers.CheckError(err)
}

func (d *Driver) conn() Conn {
	if d.transaction == nil {
		return d.connection
	}
	return d.transaction
}

func (d *Driver) QueryPasses(sql string) (bool, bool) {
	rows, err := d.conn().Query(sql)
	return err == nil, err == nil && rows.Next()
}

func (d *Driver) ExecuteBatch(sqls []string) error {
	if d.Engine == "postgres" {
		_, err := d.conn().Exec(strings.Join(sqls, ""))
		return err
	} else {
		for _, sql := range sqls {
			if _, err := d.conn().Exec(sql); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Driver) Close() {
	if d.IsDryRun {
		log.Println("Rolling back transaction...")
		d.RollbackTransaction()
	}
	log.Printf("Closing %s driver...\n", d.Engine)
	if d.connection != nil {
		err := d.connection.Close()
		helpers.CheckWarn(err)
	}
	log.Printf("Closed %s driver.\n", d.Engine)
}

func getMariaConnection(d Driver) *sql.DB {
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true", d.Username, d.Password, d.Host, d.Port, d.Schema)
	if len(d.Options) > 0 {
		url = fmt.Sprintf("%s&%s", url, strings.Join(d.Options, "&"))
	}
	db, err := sql.Open("mysql", url)
	helpers.CheckError(err)
	return db
}

func getPostgresConnection(d Driver) *sql.DB {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", d.Username, d.Password, d.Host, d.Port, d.Schema)
	if len(d.Options) > 0 {
		url = fmt.Sprintf("%s?%s", url, strings.Join(d.Options, "&"))
	}
	db, err := sql.Open("postgres", url)
	helpers.CheckError(err)
	return db
}
