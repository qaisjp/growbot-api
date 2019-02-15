package config

import "github.com/google/uuid"

type Config struct {
	LogLevel string `default:"debug"`

	Database DatabaseConfig

	// Bind Address
	BindAddress string `default:"0.0.0.0:8080"`

	// Static Robot UUID (stage 1 only)
	UUID uuid.UUID `required:"true"`
}

type DatabaseConfig struct {
	ConnectionString string `required:"true"`
}
