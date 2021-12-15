package config

import (
	"fmt"
	"os"

	"github.com/fahmifan/smol/internal/model/models"
	"github.com/joho/godotenv"
)

func init() {
	models.LogErr(godotenv.Load(".env"))
}

func mustLookupEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		models.PanicErr(fmt.Errorf("env not found %s", key))
	}
	return val
}

func Port() int {
	if port := models.StringToInt(os.Getenv("PORT")); port >= 80 {
		return port
	}
	return 8000
}

func GoogleClientID() string {
	return mustLookupEnv("GOOGLE_CLIENT_ID")
}

func GoogleClientSecret() string {
	return mustLookupEnv("GOOGLE_CLIENT_SECRET")
}

func ServerBaseURL() string {
	if baseURL := os.Getenv("SERVER_BASE_URL"); baseURL != "" {
		return baseURL
	}
	return fmt.Sprintf("http://localhost:%d", Port())
}
