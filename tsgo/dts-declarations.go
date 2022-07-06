package tsgo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func WriteFileDecls(s *strings.Builder, file *ast.File) error {
	if file == nil {
		return fmt.Errorf("Recieved nil file")
	}

	for _, decl := range file.Decls {
		err := WriteDecl(s, decl)
		if err != nil {
			return fmt.Errorf("Error writing declaration: %v", err)
		}
	}

	return nil
}

func WriteDecl(s *strings.Builder, decl ast.Decl) error {
	if decl == nil {
		return fmt.Errorf("Recieved nil declaration")
	}

	switch decl := decl.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok {
				if decl.Tok == token.CONST {
					err := WriteConstDecl(s, vs)
					if err != nil {
						return fmt.Errorf("Error declaring constant: %v", err)
					}
				} else if decl.Tok == token.VAR {
					err := WriteVarDecl(s, vs)
					if err != nil  {
						return fmt.Errorf("Error declaring variable: %v", err)
					}
				} else {
					return fmt.Errorf("Unknown value spec token: %v", decl.Tok)
				}
			} else if ts, ok := spec.(*ast.TypeSpec); ok {
				err := WriteTypeDecl(s, ts)
				if err != nil {
					return fmt.Errorf("Error declaring type alias: %v", err)
				}
			} else {
				return fmt.Errorf("Unknown spec type")
			}
		}
	case *ast.FuncDecl:
		err := WriteFuncDecl(s, decl)
		if err != nil {
			return fmt.Errorf("Error declaring function: %v", err)
		}
	default:
		return fmt.Errorf("Unknown declaration type: %t", decl)
	}

	return nil
}

func WriteVarDecl(s *strings.Builder, spec *ast.ValueSpec) error {
	if spec == nil {
		return fmt.Errorf("Recieved nil value spec")
	}

	s.WriteString("declare let ")

	var typeStr string
	if spec.Type != nil {
		tmpS := new(strings.Builder)
		err := writeTypeof(s, spec.Type)
		if err != nil {
			return fmt.Errorf("Error writing variable type: %v", err)
		}
		typeStr = tmpS.String()
	} else {
		return fmt.Errorf("Untyped variable declarations are unsupported")
	}

	for i, name := range spec.Names {
		s.WriteString(name.Name + ": " + typeStr)
		if i < len(spec.Names) - 1 {
			s.WriteString(", ")
		}
	}

	s.WriteString(";\n")
	return nil
}

func WriteConstDecl(s *strings.Builder, spec *ast.ValueSpec) error {
	if spec == nil {
		return fmt.Errorf("Recieved nil value spec")
	}

	s.WriteString("declare const ")

	var typeStr string
	if spec.Type != nil {
		tmpS := new(strings.Builder)
		err := writeTypeof(tmpS, spec.Type)
		if err != nil {
			return fmt.Errorf("Error writing type: %v", err)
		}
		typeStr = tmpS.String()
	}

	for i, name := range spec.Names {
		s.WriteString(name.Name)
		if typeStr != "" {
			s.WriteString(": " + typeStr)
		}
		s.WriteString(" = ")
		err := writeConstantValue(s, spec.Values[i])
		if err != nil {
			return fmt.Errorf("Error writing value: %v", err)
		}

		if i < len(spec.Names) - 1 {
			s.WriteString(", ")
		}
	}

	s.WriteString(";\n")
	return nil
}

func WriteFuncDecl(s *strings.Builder, fn *ast.FuncDecl) error {
	if fn == nil {
		return fmt.Errorf("Recieved nil func declaration")
	}

	s.WriteString("declare const " + fn.Name.Name + ": ")
	err := writeTypeofFunc(s, fn.Type)
	if err != nil {
		return fmt.Errorf("Error writing type: %v", err)
	}

	s.WriteString(";\n")
	return nil
}

func WriteTypeDecl(s *strings.Builder, spec *ast.TypeSpec) error {
	if spec == nil {
		return fmt.Errorf("Recieved nil file")
	}

	s.WriteString("declare type " + spec.Name.Name)
	err := writeTypeParams(s, spec.TypeParams)
	if err != nil {
		return fmt.Errorf("Error writing type params: %v", err)
	}

	s.WriteString(" = ")
	err = writeTypeof(s, spec.Type)
	if err != nil {
		return fmt.Errorf("Error writing type: %v", err)
	}

	s.WriteString(";\n")
	return nil
}