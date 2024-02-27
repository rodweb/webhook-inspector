package config

type Config struct {
	// Environment variables
	Token  string
	Domain string
	Port   string

	// Arguments
	EnableFakeRequests bool
}
