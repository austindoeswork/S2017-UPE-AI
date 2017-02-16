/*

Package config is used to define and store various server configuration
parameters.

*/

package config

import (
	"encoding/json"
)

// Config stores server configuration
type Config struct {
	*config
}

//TODO add static and template dirs to config
type config struct {
	ServerAddress    string // ServerAddress is ths address that this server starts up as.
	DatabaseAddress  string // DatabaseAddress is the address of the db server.
	DatabaseName     string // DatabaseName is the name of the database to use.
	DatabaseUsername string
	DatabasePassword string
}

// Example returns an example config
func Example() string {
	c := &Config{
		&config{
			ServerAddress:    ":8080",
			DatabaseAddress:  "tcp(127.0.0.1:3306)",
			DatabaseName:     "aicomp",
			DatabaseUsername: "root",
			DatabasePassword: "",
		},
	}
	b, _ := json.MarshalIndent(&c, "", "  ")
	return string(b)
}

// Validate validates a config
func (c *Config) Validate() error {
	if c.config == nil {
		return &ErrInvalidConfig{"config is empty"}
	}
	if len(c.config.ServerAddress) == 0 {
		return &ErrInvalidConfig{"unspecified server address"}
	}
	if len(c.config.DatabaseName) == 0 {
		return &ErrInvalidConfig{"unspecified database name"}
	}
	if len(c.config.DatabaseUsername) == 0 {
		return &ErrInvalidConfig{"unspecified database username"}
	}
	return nil
}

// ServerAddress returns the server address
func (c *Config) ServerAddress() string {
	if c.config == nil {
		return ""
	}
	return c.config.ServerAddress
}

// DatabaseAddress returns the database address
func (c *Config) DatabaseAddress() string {
	if c.config == nil {
		return ""
	}
	return c.config.DatabaseAddress
}

// DatabaseName returns the database name
func (c *Config) DatabaseName() string {
	if c.config == nil {
		return ""
	}
	return c.config.DatabaseName
}

// DatabaseUsername returns the database username
func (c *Config) DatabaseUsername() string {
	if c.config == nil {
		return ""
	}
	return c.config.DatabaseUsername
}

// DatabaseName returns the database name
func (c *Config) DatabasePassword() string {
	if c.config == nil {
		return ""
	}
	return c.config.DatabasePassword
}

// ErrInvalidConfig is an invalid config error
type ErrInvalidConfig struct {
	Reason string
}

// Error returns the error string
func (err *ErrInvalidConfig) Error() string {
	return err.Reason
}
