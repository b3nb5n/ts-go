package tsgo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// Writes the ts equivilent of the given type expression to s
func WriteTypeof(s *strings.Builder, expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.Ident:
		WriteTypeofIdent(s, expr)
	case *ast.StarExpr:
		return WriteTypeofPointer(s, expr)
	case *ast.UnaryExpr:
		return WriteTypeofUnaryExpr(s, expr)
	case *ast.BinaryExpr:
		return WriteTypeofBinaryExpr(s, expr)
	case *ast.ArrayType:
		return WriteTypeofArray(s, expr)
	case *ast.MapType:
		return WriteTypeofMap(s, expr)
	case *ast.StructType:
		return WriteTypeofStruct(s, expr)
	case *ast.InterfaceType:
		return WriteTypeofInterface(s, expr)
	case *ast.FuncType:
		return WriteTypeofFunc(s, expr)
	default:
		s.WriteString("unknown")
		return fmt.Errorf("Unrecognized type expression")
	}

	return nil
}

func WriteTypeofIdent(s *strings.Builder, ident *ast.Ident) {
	switch ident.Name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "byte", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128":
		s.WriteString("number")
	case "bool":
		s.WriteString("boolean")
	case "rune", "string":
		s.WriteString("string")
	case "error":
		s.WriteString("Error")
	default:
		s.WriteString(ident.Name)
	}
}

func WriteTypeofPointer(s *strings.Builder, ptrType *ast.StarExpr) error {
	err := WriteTypeof(s, ptrType.X)
	if err != nil {
		return fmt.Errorf("Error resolving value type: %v", err)
	}

	s.WriteString(" | undefined")
	return nil
}

func WriteTypeofArray(s *strings.Builder, arrayType *ast.ArrayType) error {
	err := WriteTypeof(s, arrayType.Elt)
	if err != nil {
		return fmt.Errorf("Error resolving element type: %v", err)
	}

	s.WriteString("[]")
	return nil
}

func WriteTypeofMap(s *strings.Builder, mapType *ast.MapType) error {
	s.WriteString("{ [key: ")
	err := WriteTypeof(s, mapType.Key)
	if err != nil {
		return fmt.Errorf("Error resolving key type: %v", err)
	}

	s.WriteString("]: ")
	err = WriteTypeof(s, mapType.Value)
	if err != nil {
		return fmt.Errorf("Error resolving value type: %v", err)
	}

	s.WriteString(" }")
	return nil
}

func WriteTypeofStruct(s *strings.Builder, structType *ast.StructType) error {
	return writeBlockType(s, structType.Fields)
}

func WriteTypeofInterface(s *strings.Builder, interfaceType *ast.InterfaceType) error {
	if numFields := interfaceType.Methods.NumFields(); numFields == 0 {
		s.WriteString("any")
		return nil
	}

	return writeBlockType(s, interfaceType.Methods)
}

func WriteTypeofFunc(s *strings.Builder, funcType *ast.FuncType) error {
	s.WriteByte('(')
	err := writeNamedFields(s, funcType.Params)
	if err != nil {
		return fmt.Errorf("Error writing params: %v", err)
	}
	s.WriteByte(')')

	if funcType.TypeParams != nil && funcType.TypeParams.NumFields() > 0 {
		s.WriteByte('<')
		err = writeFieldList(s, funcType.TypeParams, func(name, t string) { s.WriteString(name + " extends " + t) })
		if err != nil {
			return fmt.Errorf("Error writing type params: %v", err)
		}
		s.WriteByte('>')
	}

	s.WriteString(" => ")
	if numResults := funcType.Results.NumFields(); numResults == 0 {
		s.WriteString("void")
	} else if numResults == 1 {
		err := WriteTypeof(s, funcType.Results.List[0].Type)
		if err != nil {
			return fmt.Errorf("Error resolving result type: %v", err)
		}
	} else {
		if hasFieldNames(funcType.Results) {
			s.WriteString("{ ")
			err = writeNamedFields(s, funcType.Results)
			s.WriteString(" }")
		} else {
			s.WriteByte('[')
			err = writeUnnamedFields(s, funcType.Results)
			s.WriteByte(']')
		}

		if err != nil {
			return fmt.Errorf("Error writing result type: %v", err)
		}
	}

	return nil
}

func WriteTypeofUnaryExpr(s *strings.Builder, expr *ast.UnaryExpr) error {
	if expr.Op != token.TILDE {
		return fmt.Errorf("Unrecognized unary type operator: %v", expr.Op)
	}

	return WriteTypeof(s, expr.X)
}

func WriteTypeofBinaryExpr(s *strings.Builder, expr *ast.BinaryExpr) error {
	if expr.Op != token.OR {
		return fmt.Errorf("Unrecognized binary type operator: %v", expr.Op)
	}

	err := WriteTypeof(s, expr.X)
	if err != nil {
		return fmt.Errorf("Error writing lhs of binary type expression: %v", err)
	}

	s.WriteString(" | ")
	err = WriteTypeof(s, expr.Y)
	if err != nil {
		return fmt.Errorf("Error writing rhs of binary type expression: %v", err)
	}

	return nil
}

func writeFieldList(
	s *strings.Builder,
	fields *ast.FieldList,
	WriteType func(name, t string),
) error {
	numFields := fields.NumFields()
	if numFields == 0 {
		return nil
	}

	var i int
	for _, field := range fields.List {
		tmpS := new(strings.Builder)
		err := WriteTypeof(tmpS, field.Type)
		if err != nil {
			return fmt.Errorf("Error resolving field type: %v", err)
		}
		fieldType := tmpS.String()

		if field.Names == nil || len(field.Names) == 0 {
			WriteType("", fieldType)
			if i < numFields-1 {
				s.WriteString(", ")
			}

			i++
			continue
		}

		for _, name := range field.Names {
			WriteType(name.Name, fieldType)
			if i < numFields-1 {
				s.WriteString(", ")
			}

			i++
		}
	}

	return nil
}

func writeUnnamedFields(s *strings.Builder, fields *ast.FieldList) error {
	return writeFieldList(s, fields, func(_, t string) { s.WriteString(t) })
}

func writeNamedFields(s *strings.Builder, fields *ast.FieldList) error {
	var unNamedFields int
	return writeFieldList(s, fields, func(name, t string) {
		if name == "" {
			unNamedFields++
			name = strings.Repeat("_", unNamedFields)
		}

		s.WriteString(name + ": " + t)
	})
}

func writeTypeParams(s *strings.Builder, params *ast.FieldList) error {
	if params != nil && len(params.List) > 0 {
		s.WriteByte('<')
		err := writeFieldList(s, params, func(name, t string) { s.WriteString(name + " extends " + t) })
		if err != nil {
			return err
		}
		s.WriteByte('>')
	}

	return nil
}

func writeBlockType(
	s *strings.Builder,
	fields *ast.FieldList,
) error {
	
	s.WriteString("{ ")
	intersections := make([]ast.Expr, 0)
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			intersections = append(intersections, field.Type)
			continue
		}

		tmpS := new(strings.Builder)
		err := WriteTypeof(tmpS, field.Type)
		if err != nil {
			return fmt.Errorf("Error resolving field type: %v", err)
		}
		valueType := tmpS.String()

		for _, name := range field.Names {
			key := name.Name
			if !validJsIdent(key) {
				key = "'" + key + "'"
			}

			s.WriteString(key + ": " + valueType + "; ")
		}
	}

	s.WriteByte('}')

	for _, intersection := range intersections {
		s.WriteString(" & (")
		err := WriteTypeof(s, intersection)
		if err != nil {
			return fmt.Errorf("Error writing intersection type member: %v", err)
		}
		s.WriteByte(')')
	}

	return nil
}
