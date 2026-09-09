package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr                                        string
	PostgresDSN, Neo4jURI, Neo4jUser, Neo4jPassword string
	GitHubMaxBytes                                  int64
	GitHubTimeout                                   time.Duration
}

func Load() Config {
	n := int64(10 << 20)
	if x, e := strconv.ParseInt(os.Getenv("GITHUB_MAX_DOWNLOAD_BYTES"), 10, 64); e == nil && x > 0 {
		n = x
	}
	t := 20
	if x, e := strconv.Atoi(os.Getenv("GITHUB_TIMEOUT_SECONDS")); e == nil && x > 0 {
		t = x
	}
	return Config{HTTPAddr: os.Getenv("HTTP_ADDR"), PostgresDSN: os.Getenv("POSTGRES_DSN"), Neo4jURI: os.Getenv("NEO4J_URI"), Neo4jUser: os.Getenv("NEO4J_USER"), Neo4jPassword: os.Getenv("NEO4J_PASSWORD"), GitHubMaxBytes: n, GitHubTimeout: time.Duration(t) * time.Second}
}
