package task

type Config struct {
	// Node Name
	NodeName string
}

func NewConfig(name string) *Config {
	return &Config{
		NodeName: name,
	}
}
