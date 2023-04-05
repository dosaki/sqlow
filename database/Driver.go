package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"log"
	"sqlow/helpers"
	"strings"
)

type Driver struct {
	Engine     string
	Host       string
	Port       string
	Schema     string
	Username   string
	Password   string
	Options    []string
	connection *sql.DB
}

func (d *Driver) Connect() {
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
}

func (d *Driver) QueryPasses(sql string) (bool, bool) {
	rows, err := d.connection.Query(sql)
	return err == nil, err == nil && rows.Next()
}

func (d *Driver) ExecuteBatch(sqls []string) error {
	if d.Engine == "postgres" {
		_, err := d.connection.Exec(strings.Join(sqls, ""))
		return err
	} else {
		for _, sql := range sqls {
			if _, err := d.connection.Exec(sql); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Driver) Close() {
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
