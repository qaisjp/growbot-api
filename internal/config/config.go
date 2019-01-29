package config

type Config struct {
	LogLevel string `default:"debug"`

	// Bind Address
	BindAddress string `default:"0.0.0.0:8080"`
}
