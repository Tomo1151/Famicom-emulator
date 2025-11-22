package config

// Config holds emulator runtime options shared across windows.
type Config struct {
	ScaleFactor int
	SoundVolume float32
	ShowFPS     bool
}

// New returns a Config initialized with default values.
func New() *Config {
	return &Config{
		ScaleFactor: 3,
		SoundVolume: 1.0,
		ShowFPS:     false,
	}
}
