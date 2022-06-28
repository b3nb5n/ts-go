package tsgo

import (
	"go/ast"
	"go/token"
	"strings"
)

func (gen *Generator) WriteGenDecl(s *strings.Builder, decl *ast.GenDecl) {
	state := &groupDeclState{iotaValue: -1}
	for _, spec := range decl.Specs {
		gen.writeSpec(s, spec, state)
		s.WriteByte('\n')
	}

	s.WriteByte('\n')
}

func (gen *Generator) WriteFuncDecl(s *strings.Builder, decl *ast.FuncDecl) {
	s.WriteString("type ")
	s.WriteString(decl.Name.Name)
	if decl.Type.TypeParams != nil {
		gen.writeTypeParams(s, decl.Type.TypeParams.List)
	}

	s.WriteString(" = (")

	var i int
	last := decl.Type.Params.NumFields() - 1
	for _, param := range decl.Type.Params.List {
		tmpS := new(strings.Builder)
		gen.writeExpr(tmpS, param.Type, 0)
		typeStr := tmpS.String()

		for _, name := range param.Names {
			s.WriteString(name.Name)
			s.WriteString(": ")
			s.WriteString(typeStr)
			if i < last {
				s.WriteString(", ")
			}

			i++
		}
	}

	s.WriteString(") => ")
	if decl.Type.Results == nil || len(decl.Type.Results.List) == 0 {
		s.WriteString("void")
	} else if len(decl.Type.Results.List) == 1 && len(decl.Type.Results.List[0].Names) <= 1 {
		gen.writeExpr(s, decl.Type.Results.List[0].Type, 0)
	} else {
		s.WriteByte('[')

		var i int
		last := decl.Type.Results.NumFields() - 1
		for _, result := range decl.Type.Results.List {
			tmpS := new(strings.Builder)
			gen.writeExpr(tmpS, result.Type, 0)
			typeStr := tmpS.String()

			if len(result.Names) > 0 {
				for range result.Names {
				s.WriteString(typeStr)
				if i < last {
					s.WriteString(", ")
				}

				i++
			}
			} else {
				s.WriteString(typeStr)
				if i < last {
					s.WriteString(", ")
				}

				i++
			}
		}

		s.WriteByte(']')
	}
}

func (gen *Generator) WriteFile(s *strings.Builder, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {

		// GenDecl can be an import, type, var, or const expression
		case *ast.GenDecl:
			if tok := node.Tok; tok == token.VAR || tok == token.IMPORT {
				return false
			}

			gen.WriteGenDecl(s, node)
			return false
		case *ast.FuncDecl:
			gen.WriteFuncDecl(s, node)
			return false
		}

		return true
	})
}
