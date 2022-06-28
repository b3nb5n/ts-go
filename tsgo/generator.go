package tsgo

type Generator struct {
	config *Config
}

func NewGenerator(config *Config) *Generator {
	gen := &Generator{
		config: config,
	}

	if config == nil {
		gen.config = NewConfig()
	}

	return gen
}
