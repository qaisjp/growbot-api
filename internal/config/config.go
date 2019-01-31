package config

import "github.com/google/uuid"

type Config struct {
	LogLevel string `default:"debug"`

	// Bind Address
	BindAddress string `default:"0.0.0.0:8080"`

	// Static Robot UUID (stage 1 only)
	UUID uuid.UUID `required:"true"`
}
