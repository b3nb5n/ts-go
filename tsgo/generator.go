package tsgo

// type GroupedDeclarationMode int
// const (
// 	// Declare each of the go values as a sperate ts constant
// 	Const GroupedDeclarationMode = iota

// 	// Declare a ts enum with entries for each of the go values that reference the initial iota value
// 	IotaEnum

// 	// Declare a ts enum with entries for each of the go values
// 	Enum
// )

type Config struct {
	// // Customize the indentation (use \t if you want tabs)
	// Indent string

	// // Determines how the generator should treat grouped declarations
	// GroupedDeclarationMode GroupedDeclarationMode

	// Specify your own custom type translations, useful for custom types, `time.Time` and `null.String`.
	// Be default unrecognized types will be output as `any /* name */`.
	TypeMappings map[string]string
}

func NewConfig() *Config {
	return &Config{
		TypeMappings: make(map[string]string),
	}
}

type Generator struct {
	config *Config
}

func NewGenerator(config *Config) *Generator {
	if config == nil {
		config = NewConfig()
	}

	return &Generator{ config: config }
}
