package migrations

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"sqlow/database"
	"sqlow/domains"
	"sqlow/helpers"
)

func ResolveMigrations(filePath string) []domains.Migration {
	dir := filepath.Dir(filePath)
	log.Printf("Resolving migrations in %s...\n", filePath)
	content, err := os.ReadFile(filePath)
	helpers.CheckWarn(err)
	m := make(map[string][]domains.Migration)
	err = yaml.Unmarshal(content, &m)
	helpers.CheckWarn(err)
	var migrations []domains.Migration
	for _, migration := range m["migrations"] {
		migration.ResolveFiles(dir)
		migrations = append(migrations, migration)
	}
	return migrations
}

func RunMigrations(migrations []domains.Migration, driver database.Driver) {
	for _, migration := range migrations {
		log.Printf("  - %s\n", migration.Description)
		migration.Run(driver)
	}
}
