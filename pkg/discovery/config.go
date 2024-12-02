package discovery

import (
	"time"
)

// Option is a function type that modifies Config
type Option func(*Config)

// Config holds the configuration for the service discovery node
type Config struct {
	EnableDHT          bool
	EnablePubSub       bool
	EnablePeerExchange bool
	PeerTTL            time.Duration
	Options            []Option
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		EnableDHT:          true,
		EnablePubSub:       true,
		EnablePeerExchange: true,
		PeerTTL:            3 * time.Hour,
		Options:            []Option{},
	}
}

// WithDHT enables or disables DHT
func WithDHT(enable bool) Option {
	return func(c *Config) {
		c.EnableDHT = enable
	}
}

// WithPubSub enables or disables PubSub
func WithPubSub(enable bool) Option {
	return func(c *Config) {
		c.EnablePubSub = enable
	}
}

// WithPeerTTL sets the peer time-to-live duration
func WithPeerTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.PeerTTL = ttl
	}
}

// WithPeerExchange enables or disables peer exchange service
func WithPeerExchange(enable bool) Option {
	return func(c *Config) {
		c.EnablePeerExchange = enable
	}
}
