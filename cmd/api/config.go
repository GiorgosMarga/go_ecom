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
		mongoURI:  readENV("ECOMGO_URI", "mongodb://ecomgo_catbreathe:a950d57379fe11347d4f762a099740753a193eea@nor.h.filess.io:27018/ecomgo_catbreathe"),
		jwtSecret: []byte(readENV("JWT_SECRET", "password")),
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
