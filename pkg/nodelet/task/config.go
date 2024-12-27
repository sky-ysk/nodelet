package task

type Config struct {
	// Node ID
	name string
}

func NewConfig(name string) *Config {
	return &Config{
		name: name,
	}
}
