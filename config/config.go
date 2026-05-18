package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                            string
	DatabaseURL                     string
	JWTPublicKey                    string
	AllowedOrigins                  []string
	TracerURL                       string
	TracerServiceKey                string
	PubSubProjectID                 string
	PubSubAudience                  string
	PubSubJobCreatedTopicID         string
	PubSubInterviewScheduledTopicID string
	PubSubRubeOwidTopicID           string
}

func parseOrigins(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func Load() *Config {
	if _, err := os.Stat(".env.local"); err == nil {
		_ = godotenv.Load(".env.local")
	}

	return &Config{
		Port:                            os.Getenv("PORT"),
		DatabaseURL:                     os.Getenv("DATABASE_URL"),
		JWTPublicKey:                    os.Getenv("JWT_PUBLIC_KEY"),
		AllowedOrigins:                  parseOrigins(os.Getenv("ALLOWED_ORIGINS")),
		TracerURL:                       os.Getenv("TRACER_SERVICE_URL"),
		TracerServiceKey:                os.Getenv("TRACER_SERVICE_KEY"),
		PubSubProjectID:                 os.Getenv("PUBSUB_PROJECT_ID"),
		PubSubAudience:                  os.Getenv("PUBSUB_AUDIENCE"),
		PubSubJobCreatedTopicID:         "job-created",
		PubSubInterviewScheduledTopicID: "job-interview-scheduled",
		PubSubRubeOwidTopicID:           "rube-owid",
	}
}
