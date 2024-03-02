package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbUser     string
	dbPassword string
	dbName     string
	dbHost     = "localhost"
	dbPort     = "3306"
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
		log.Fatal("DB_NAME environment variable is not set")
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		dbHost = v
	} else {
		log.Printf("DB_HOST environment variable is not set, using default value (%s)\n", dbHost)
	}

	if v := os.Getenv("DB_PORT"); v != "" {
		dbPort = v
	} else {
		log.Printf("DB_PORT environment variable is not set, using default value (%s)\n", dbPort)
	}
}

func main() {
	load_env_vars()

	// Build the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping the database: %v", err)
	}

	migrationFiles := []string{
		"users.sql",
	}

	for _, file := range migrationFiles {
		if err := executeSQLFile(db, file); err != nil {
			log.Fatalf("Failed to execute migration file (%s): %v", file, err)
		}
		log.Printf("Successfully executed migration file: %s", file)
	}
}
