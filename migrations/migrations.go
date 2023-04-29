package migrations

import (
	"fmt"
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
	var sqlToRun []string
	for _, migration := range migrations {
		migrationSQL := migration.Collect(driver)
		if len(migrationSQL) > 0 {
			if driver.IsDryRun {
				sqlToRun = append(sqlToRun, fmt.Sprintf("\n-- %s", migration.Description))
			} else {
				log.Printf("  - %s\n", migration.Description)
			}
			sqlToRun = append(sqlToRun, migrationSQL...)
		}
		err := driver.ExecuteBatch(migrationSQL)
		helpers.CheckError(err)
	}

	if driver.IsDryRun {
		for _, sql := range sqlToRun {
			fmt.Println(sql)
		}
	}
}
