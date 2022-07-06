package tsgo

import (
	"fmt"
	"go/ast"
	"strings"
)

// returns the typescript equivilent of the given type
func writeTypeof(s *strings.Builder, expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.Ident:
		writeTypeofIdent(s, expr)
	case *ast.StarExpr:
		return writeTypeofPointer(s, expr)
	case *ast.ArrayType:
		return writeTypeofArray(s, expr)
	case *ast.MapType:
		return writeTypeofMap(s, expr)
	case *ast.StructType:
		return writeTypeofStruct(s, expr)
	case *ast.FuncType:
		return writeTypeofFunc(s, expr)
	default:
		s.WriteString("unknown")
		return fmt.Errorf("Unrecognized type expression")
	}

	return nil
}

// func writeTypeofConstantValue(s *strings.Builder, value ast.Expr) error {
// 	switch value := value.(type) {
// 	case *ast.Ident:
// 		switch value.Name {
// 		case "true", "false":
			
// 		}
// 	case *ast.BasicLit:
// 		switch value.Kind {
// 		case token.INT, token.FLOAT, token.IMAG:
// 			s.WriteString("number")
// 		case token.CHAR, token.STRING:
// 			s.WriteString("string")
// 		}
// 	}
// }

func writeTypeofIdent(s *strings.Builder, ident *ast.Ident) {
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

func writeTypeofPointer(s *strings.Builder, ptrType *ast.StarExpr) error {
	err := writeTypeof(s, ptrType.X)
	if err != nil {
		return fmt.Errorf("Error resolving value type: %v", err)
	}

	s.WriteString(" | undefined")
	return nil
}

func writeTypeofArray(s *strings.Builder, arrayType *ast.ArrayType) error {
	err := writeTypeof(s, arrayType.Elt)
	if err != nil {
		return fmt.Errorf("Error resolving element type: %v", err)
	}
	
	s.WriteString("[]")
	return nil
}

func writeTypeofMap(s *strings.Builder, mapType *ast.MapType) error {
	s.WriteString("{ [key: ")
	err := writeTypeof(s, mapType.Key)
	if err != nil {
		return fmt.Errorf("Error resolving key type: %v", err)
	}

	s.WriteString("]: ")
	err = writeTypeof(s, mapType.Value)
	if err != nil {
		return fmt.Errorf("Error resolving value type: %v", err)
	}

	s.WriteString(" }")
	return nil
}

func writeTypeofStruct(s *strings.Builder, structType *ast.StructType) error {
	s.WriteString("{ ")

	var i int
	numFields := structType.Fields.NumFields()
	for _, field := range structType.Fields.List {
		tmpS := new(strings.Builder)
		err := writeTypeof(tmpS, field.Type)
		if err != nil {
			return fmt.Errorf("Error resolving field type: %v", err)
		}
		valueType := tmpS.String()

		for _, name := range field.Names {
			key := name.Name
			if !validJsIdent(key) {
				key = "'" + key + "'"
			}

			s.WriteString(key + ": " + valueType)
			if i < numFields - 1 {
				s.WriteString(";")
			}
			s.WriteByte(' ')

			i++
		}
	}

	s.WriteByte('}')
	return nil
}

func writeTypeofInterface(s *strings.Builder, interfaceType *ast.InterfaceType) error {
	s.WriteString("{ ")
	for _, field := range interfaceType.Methods.List {
		tmpS := new(strings.Builder)
		err := writeTypeof(tmpS, field.Type)
		if err != nil {
			return fmt.Errorf("Error resolving field type: %v", err)
		}
		fieldType := tmpS.String()

		for _, name := range field.Names {
			key := name.Name
			if !validJsIdent(key) {
				key = "'" + key + "'"
			}

			s.WriteString(key + ": " + fieldType + "; ")
		}
	}

	s.WriteByte('}')
	return nil
}

func writeTypeofFunc(s *strings.Builder, funcType *ast.FuncType) error {
	s.WriteByte('(')
	err := writeNamedFields(s, funcType.Params)
	if err != nil {
		return fmt.Errorf("Error writing params: %v", err)
	}
	s.WriteByte(')')

	if funcType.TypeParams != nil && funcType.TypeParams.NumFields() > 0 {
		s.WriteByte('<')
		err = writeFieldList(s, funcType.TypeParams, func(name, t string) {s.WriteString(name + " extends " + t)})
		if err != nil {
			return fmt.Errorf("Error writing type params: %v", err)
		}
		s.WriteByte('>')
	}

	s.WriteString(" => ")
	if numResults := funcType.Results.NumFields(); numResults == 0 {
		s.WriteString("void")
	} else if numResults == 1 {
		err := writeTypeof(s, funcType.Results.List[0].Type)
		if err != nil {
			return fmt.Errorf("Error resolving result type: %v", err)
		}
	} else {
		if hasFieldNames(funcType.Results) {
			s.WriteByte('{')
			err = writeNamedFields(s, funcType.Results)
			s.WriteByte('}')
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

func writeFieldList(
	s *strings.Builder,
	fields *ast.FieldList,
	writeType func(name, t string),
) error {
	numFields := fields.NumFields()
	if numFields == 0 {
		return nil
	}

	var i int
	for _, field := range fields.List {
		tmpS := new(strings.Builder)
		err := writeTypeof(tmpS, field.Type)
		if err != nil {
			return fmt.Errorf("Error resolving field type: %v", err)
		}
		fieldType := tmpS.String()

		if field.Names == nil || len(field.Names) == 0 {
			writeType("", fieldType)
			if i < numFields - 1 {
				s.WriteString(", ")
			}

			i++
			continue
		}

		for _, name := range field.Names {
			writeType(name.Name, fieldType)
			if i < numFields - 1 {
				s.WriteString(", ")
			}

			i++
		}
	}

	return nil
}

func writeUnnamedFields(s *strings.Builder, fields *ast.FieldList) error {
	return writeFieldList(s, fields, func(_, t string) {s.WriteString(t)})
}

func writeNamedFields(s *strings.Builder, fields *ast.FieldList) error {
	return writeFieldList(s, fields, func(name, t string) {s.WriteString(name + ": " + t)})
}

func writeTypeParams(s *strings.Builder, params *ast.FieldList) error {
	if params != nil && len(params.List) > 0 {
		s.WriteByte('<')
		err := writeFieldList(s, params, func(name, t string) {s.WriteString(name + " extends " + t)})
		if err != nil {
			return err
		}
		s.WriteByte('>')
	}

	return nil
}
