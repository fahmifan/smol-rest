package config

import (
	"fmt"
	"os"

	"github.com/fahmifan/smol/backend/model/models"
)

func Port() int {
	if port := models.StringToInt(os.Getenv("PORT")); port >= 80 {
		return port
	}
	return 8000
}

func MustLookupEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		models.PanicErr(fmt.Errorf("env not found %s", key))
	}
	return val
}
