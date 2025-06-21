package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadDotenvIfDebug loads .env for local debugging.
// exists program if DEBUG_TASK_MGMT is set to "true" while the .env file is missing
func LoadDotenvIfDebug() bool {
	err := godotenv.Load()
	debug, okDebug := os.LookupEnv("DEBUG_TASK_MGMT")

	if okDebug && strings.ToLower(debug) == "true" {
		if err!=nil{
			log.Fatal(".env file not found or failed to load")
		}
		return true
	}
	return false
}
