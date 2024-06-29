package config

import (
	"net/http"
	"time"
)

type ServerConfig struct {
	Port           string
	Handler        http.Handler
	MaxHeaderBytes int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}
