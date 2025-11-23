package dotenv

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// get from system env first then if did not find it, try .env on project
// for container solution

var (
	loadOnce sync.Once
)

func ensureLoaded() {
	_ = godotenv.Load()
}

func Env(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	loadOnce.Do(ensureLoaded)
	return os.Getenv(key)
}
