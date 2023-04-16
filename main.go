package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"log"
	"os"
	"sqlow/config"
	"sqlow/io"
	"sqlow/migrations"
	"strings"
)

func main() {
	usage := "sqlow (version: " + config.VERSION + `)

Usage:
  sqlow [options] -p<password> run <path>
  sqlow --version
  sqlow --help

Options:
  -h --help                 Show this Help.
  -r --recursive            Traverse the directory structure for .yml or .yaml files.
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
		fmt.Println("sqlow " + config.VERSION)
		os.Exit(0)
	}

	path, _ := arguments.String("<path>")
	isRecursive, _ := arguments.Bool("--recursive")
	configPath, _ := arguments.String("--config")
	engine := arguments["--engine"]
	host := arguments["--host"]
	port := arguments["--port"]
	schema := arguments["--schema"]
	username := arguments["--username"]
	password := arguments["--password"]
	options := arguments["--options"]

	log.Println("Started.")
	config := config.FromYAML(configPath).WithOverrides(engine, host, port, schema, username, password, options)

	driver := config.MakeDriver()
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
