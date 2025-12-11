package configure

// DefaultConfig defines default notification settings (e.g., GitHub Actions stderr)
type DefaultConfig struct {
	Enabled bool `yaml:"enabled,omitempty"`
}
