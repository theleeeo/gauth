package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/theleeeo/thor/runner"
	"gopkg.in/yaml.v3"
)

var (
	dbUser     string
	dbPassword string
	dbName     = "thor"
	dbAddr     = "localhost:3306"
)

func executeSQLFile(db *sql.DB, filePath string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading SQL file (%s): %v", filePath, err)
	}

	request := string(fileContent)
	_, err = db.Exec(request)
	if err != nil {
		return fmt.Errorf("error executing SQL file (%s): %v", filePath, err)
	}

	return nil
}

func load_env_vars() {
	if v := os.Getenv("DB_USER"); v != "" {
		dbUser = v
	} else {
		log.Fatal("DB_USER environment variable is not set")
	}

	if v := os.Getenv("DB_PASSWORD"); v != "" {
		dbPassword = v
	} else {
		log.Println("DB_PASSWORD environment variable is not set, using empty password")
	}

	if v := os.Getenv("DB_NAME"); v != "" {
		dbName = v
	} else {
		log.Printf("DB_NAME environment variable is not set, using default value (%s)\n", dbAddr)
	}

	if v := os.Getenv("DB_ADDR"); v != "" {
		dbAddr = v
	} else {
		log.Printf("DB_ADDR environment variable is not set, using default value (%s)\n", dbAddr)
	}
}

func loadConfig() (*runner.Config, error) {
	content, err := os.ReadFile("./.thor.yml")
	if err != nil {
		return nil, err
	}

	var config runner.Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Println("error loading config file, moving on with env-vars:", err)
		load_env_vars()
	} else {
		dbUser = cfg.RepoCfg.MySql.User
		dbPassword = cfg.RepoCfg.MySql.Password
		dbName = cfg.RepoCfg.MySql.Database
		dbAddr = cfg.RepoCfg.MySql.Addr
	}

	// Build the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		dbUser, dbPassword, dbAddr, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping the database: %v", err)
	}

	migrationFiles := []string{
		"migrations/users.sql",
		"migrations/user_providers.sql",
		"migrations/roles.sql",
		"migrations/user_roles.sql",
		"migrations/role_permissions.sql",
	}

	for _, file := range migrationFiles {
		if err := executeSQLFile(db, file); err != nil {
			log.Fatalf("Failed to execute migration file (%s): %v", file, err)
		}
		log.Printf("Successfully executed migration file: %s", file)
	}
}
