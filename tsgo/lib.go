package tsgo

import (
	"go/ast"
	"go/token"
	"unsafe"
)

type Generator struct {
	out    []byte
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

func (gen *Generator) String() string {
	// Directly ripped from the strings.Builder implementation
	return *(*string)(unsafe.Pointer(&gen.out))
}

func (gen *Generator) ParseAstNode(node ast.Node) (string, error) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch node := n.(type) {

		// GenDecl can be an import, type, var, or const expression
		case *ast.GenDecl:
			if tok := node.Tok; tok == token.VAR || tok == token.IMPORT {
				return false
			}

			gen.writeDecl(node)
			return false
		}

		return true
	})

	return "", nil
}
