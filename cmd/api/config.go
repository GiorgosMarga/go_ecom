package main

import "os"

type config struct {
	port      string
	mongoURI  string
	jwtSecret []byte
	ginMode   string
	s3AccessKey string
	s3SecretKey string
}

func NewConfig() *config {
	return &config{
		port:      readENV("PORT", "8080"),
		mongoURI:  readENV("ECOMGO_URI", ""),
		jwtSecret: []byte(readENV("JWT_SECRET", "secret")),
		ginMode:   readENV("GIN_MODE", "debug"),
		s3AccessKey: readENV("S3_ACCESS_KEY", ""),
		s3SecretKey: readENV("S3_Secret_KEY", ""),
	}
}

func readENV(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}
