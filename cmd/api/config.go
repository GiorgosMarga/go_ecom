package main

import "os"

type config struct {
	port      string
	mongoURI  string
	jwtSecret []byte
	ginMode   string
}

func NewConfig() *config {
	return &config{
		port:      readENV("PORT", "8080"),
		mongoURI:  readENV("ECOMGO_URI", ""),
		jwtSecret: []byte(readENV("JWT_SECRET", "")),
		ginMode:   readENV("GIN_MODE", "debug"),
	}
}

func readENV(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}
