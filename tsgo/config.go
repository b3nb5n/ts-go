package tsgo

type Config struct {
	// Customize the indentation (use \t if you want tabs)
	Indent string

	// Specify your own custom type translations, useful for custom types, `time.Time` and `null.String`.
	// Be default unrecognized types will be output as `any /* name */`.
	TypeMappings map[string]string
}

func NewConfig() *Config {
	return &Config{
		Indent:       "\t",
		TypeMappings: make(map[string]string),
	}
}
