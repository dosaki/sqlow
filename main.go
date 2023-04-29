package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"sqlow/config"
	"sqlow/io"
	"sqlow/migrations"
	"strings"
)

var VERSION = "development"

func main() {
	usage := "sqlow (version: " + VERSION + `)

Usage:
  sqlow [options] -p<password> run <path>
  sqlow --version
  sqlow --help

Options:
  -h --help                 Show this Help.
  -r --recursive            Traverse the directory structure for .yml or .yaml files.
  -d --dry-run              Print the non-check SQL statements that will be run (this will run the migrations, but will roll them back).
  -c --config=<configPath>  Path to config [default: ./config.yml].
  -e --engine=<engine>      Database engine (overrides config).
  -h --host=<host>          Database host (overrides config).
  -P --port=<port>          Database port (overrides config).
  -s --schema=<schema>      Database schema (overrides config).
  -u --username=<username>  Database username (overrides config).
  -p --password=<password>  Database password (required).
  -o --options=<opts>       Comma separated list of key:value options (merges with config).

Examples:
  sqlow -r -pSup4rS@f3P@ssw0rd run ./migrations
  sqlow -pSup4rS@f3P@ssw0rd run ./migrations/a_migration.yml
  sqlow -c ./my-config.yml -pSup4rS@f3P@ssw0rd run ./migrations/a_migration.yml
  sqlow -r -c ./my-config.yml -udbuser -pSup4rS@f3P@ssw0rd run ./migrations/`

	arguments, _ := docopt.ParseDoc(usage)
	version, _ := arguments.Bool("--version")
	if version {
		fmt.Println("sqlow " + VERSION)
		os.Exit(0)
	}

	path, _ := arguments.String("<path>")
	isRecursive, _ := arguments.Bool("--recursive")
	isDryRun, _ := arguments.Bool("--dry-run")
	configPath, _ := arguments.String("--config")
	engine := arguments["--engine"]
	host := arguments["--host"]
	port := arguments["--port"]
	schema := arguments["--schema"]
	username := arguments["--username"]
	password := arguments["--password"]
	options := arguments["--options"]

	log.Println("Loading config...")
	configuration := config.FromYAML(configPath).WithOverrides(engine, host, port, schema, username, password, options)
	if slices.Contains([]string{"maria", "mariadb", "mysql"}, configuration.Engine) && isDryRun {
		log.Println("MariaDB and MySQL don't support dry-runs as they can implicitly end transactions!")
		log.Println("For more information, see:")
		log.Println("* https://mariadb.com/kb/en/sql-statements-that-cause-an-implicit-commit/")
		log.Fatalln("* https://dev.mysql.com/doc/refman/8.0/en/implicit-commit.html")
	}

	log.Println("Started.")
	driver := configuration.MakeDriver(isDryRun)
	driver.Connect()
	defer driver.Close()
	files := io.GetAllFiles(path, isRecursive)
	for _, file := range files {
		if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
			migrationList := migrations.ResolveMigrations(file)
			migrations.RunMigrations(migrationList, driver)
		}
	}
	log.Println("Done!")
}
