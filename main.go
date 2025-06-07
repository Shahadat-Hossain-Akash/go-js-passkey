package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"go-js-passkey/db"
	"go-js-passkey/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading environment variables directly")
	}

	// Read env vars
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	// Initialize logger
	logInstance := initializeLogger()
	defer logInstance.Close()

	// Connect to DB
	pgDB, err := db.Connect(host, port, user, password, dbname, sslmode)
	if err != nil {
		logInstance.Error("DB connection failed", err)
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer pgDB.Close()

	http.Handle("/", http.FileServer(http.Dir("public")))

	const addr = ":8080"
	logInstance.Info("Server starting on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logInstance.Error("Server failed to start", err)
		log.Fatalf("Server failed: %v", err)
	}
}

func initializeLogger() *logger.Logger {
	logInstance, err := logger.NewLogger("go-js-passkey.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	return logInstance
}
