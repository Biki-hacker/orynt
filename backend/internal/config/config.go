package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds environment variables
type Config struct {
	Port                    string
	JWTSecret               string
	GeminiAPIKey            string
	UseFirestore            bool
	FirebaseProjectID       string
	FirebaseCredentialsFile string
	UseRedis                bool
	RedisURL                string
}

// LoadConfig reads .env file (if present) then reads env variables with defaults
func LoadConfig() *Config {
	// Load .env file silently — if it doesn't exist we fall back to OS env
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No .env file found — using system environment variables")
	} else {
		log.Println("[CONFIG] Loaded .env file successfully")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "orynt-default-secret-hackathon-2026-key"
	}

	apiKey := os.Getenv("GEMINI_API_KEY")

	useFirestore := getEnvBool("USE_FIRESTORE", false)
	firebaseProjectID := os.Getenv("FIREBASE_PROJECT_ID")
	firebaseCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	useRedis := getEnvBool("USE_REDIS", false)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Fallback to REDIS_ADDR
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
		redisURL = "redis://" + redisAddr
	}

	return &Config{
		Port:                    port,
		JWTSecret:               secret,
		GeminiAPIKey:            apiKey,
		UseFirestore:            useFirestore,
		FirebaseProjectID:       firebaseProjectID,
		FirebaseCredentialsFile: firebaseCreds,
		UseRedis:                useRedis,
		RedisURL:                redisURL,
	}
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}
