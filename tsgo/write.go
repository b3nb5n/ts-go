package tsgo

import (
	"fmt"
	"strconv"
	"strings"

	"go/ast"
	"go/token"
)

func (gen *Generator) writeIndent(s *strings.Builder, depth int) {
	s.WriteString(strings.Repeat(gen.config.Indent, depth))
}

// persists the state of a grouped declaration
// across its value specifiers
type groupDeclState struct {
	currValue string
	currType  string

	iotaValue  int
	iotaOffset int
}

func (gen *Generator) writeSpec(s *strings.Builder, _spec ast.Spec, groupState *groupDeclState) {
	switch spec := _spec.(type) {
	case *ast.TypeSpec:
		gen.writeTypeSpec(s, spec)
	case *ast.ValueSpec:
		gen.writeValueSpec(s, spec, groupState)
	}
}

// Writing of type specs, which are expressions like
// `type X struct { ... }`
// or
// `type Bar = string`
func (gen *Generator) writeTypeSpec(s *strings.Builder, ts *ast.TypeSpec) {
	if !ts.Name.IsExported() {
		return
	}

	st, isStruct := ts.Type.(*ast.StructType)
	if isStruct {
		s.WriteString("export interface ")
		s.WriteString(ts.Name.Name)
		if ts.TypeParams != nil {
			gen.writeTypeParams(s, ts.TypeParams.List)
		}

		s.WriteByte(' ')
		gen.writeStruct(s, st, 0)
	} else {
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)
		s.WriteString(" = ")

		id, isIdent := ts.Type.(*ast.Ident)
		if isIdent {
			s.WriteString(getIdent(id.Name))
		} else {
			gen.writeExpr(s, ts.Type, 0)
		}
	}
}

// Writign of value specs, which are exported const expressions like
// const SomeValue = 3
func (gen *Generator) writeValueSpec(s *strings.Builder, vs *ast.ValueSpec, group *groupDeclState) {
	if vs.Type != nil {
		tmpS := new(strings.Builder)
		gen.writeExpr(tmpS, vs.Type, 0)
		group.currType = tmpS.String()
	}

	for i, name := range vs.Names {
		group.iotaValue++
		if !name.IsExported() {
			continue
		}

		// if a value is explicitly defined
		if len(vs.Values) > i {
			tmpBuilder := new(strings.Builder)
			gen.writeExpr(tmpBuilder, vs.Values[i], 0)
			valueStr := tmpBuilder.String()

			if strings.Contains(valueStr, "iota") {
				group.currValue = "iota"
				group.iotaOffset = parseIotaOffset(valueStr)
			} else {
				group.currValue = valueStr
			}

			// if a value is specified without a type unset the group type
			if vs.Type == nil {
				group.currType = ""
			}
		}

		s.WriteString("export const ")
		s.WriteString(name.Name)

		if group.currType != "" {
			s.WriteString(": ")
			s.WriteString(group.currType)
		}

		s.WriteString(" = ")

		if group.currValue == "iota" {
			s.WriteString(strconv.Itoa(group.iotaValue + group.iotaOffset))
		} else {
			s.WriteString(group.currValue)
		}
	}
}

func (gen *Generator) writeExpr(s *strings.Builder, expr ast.Expr, depth int) {
	switch t := expr.(type) {
	case *ast.BasicLit:
		s.WriteString(t.Value)
	case *ast.Ident:
		s.WriteString(getIdent(t.String()))
	case *ast.ParenExpr:
		s.WriteByte('(')
		gen.writeExpr(s, t.X, depth)
		s.WriteByte(')')
	case *ast.StarExpr:
		gen.writeExpr(s, t.X, depth)
		s.WriteString(" | undefined")
	case *ast.SelectorExpr: // e.g. `time.Time`
		longType := fmt.Sprintf("%s.%s", t.X, t.Sel)
		mappedType, ok := gen.config.TypeMappings[longType]
		if ok {
			s.WriteString(mappedType)
		} else { // For unknown types we put `any`
			s.WriteString("any /* ")
			s.WriteString(longType)
			s.WriteString(" */")
		}
	case *ast.UnaryExpr:
		if t.Op == token.TILDE {
			// We just ignore the tilde token, in Typescript extended types are
			// put into the generic typing itself, which we can't support yet.
			gen.writeExpr(s, t.X, depth)
		} else {
			panic(fmt.Errorf("unhandled unary expr: %v\n %T", t, t))
		}
	case *ast.BinaryExpr:
		gen.writeExpr(s, t.X, depth)
		s.WriteByte(' ')
		s.WriteString(t.Op.String())
		s.WriteByte(' ')
		gen.writeExpr(s, t.Y, depth)
	case *ast.IndexExpr:
		gen.writeExpr(s, t.X, depth)
		s.WriteByte('<')
		gen.writeExpr(s, t.Index, depth)
		s.WriteByte('>')
	case *ast.IndexListExpr:
		gen.writeExpr(s, t.X, depth)
		s.WriteByte('<')
		for i, index := range t.Indices {
			gen.writeExpr(s, index, depth)
			if i != len(t.Indices)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteByte('>')
	case *ast.Ellipsis:
		s.WriteString("...")
		gen.writeExpr(s, t.Elt, depth)
		s.WriteString("[]")
	case *ast.InterfaceType:
		gen.writeInterface(s, t, depth+1)
	case *ast.StructType:
		gen.writeStruct(s, t, depth+1)
	case *ast.ArrayType:
		gen.writeArray(s, t, depth)
	case *ast.MapType:
		gen.writeMap(s, t, depth)
	case *ast.CallExpr:
		s.WriteString("any")
	default:
		panic(fmt.Errorf("unhandled: %s\n %T", t, t))
	}
}

func (gen *Generator) writeTypeParams(s *strings.Builder, fields []*ast.Field) {
	s.WriteByte('<')
	for i, f := range fields {
		for j, ident := range f.Names {
			s.WriteString(ident.Name)
			s.WriteString(" extends ")
			gen.writeExpr(s, f.Type, 0)

			if i != len(fields)-1 || j != len(f.Names)-1 {
				s.WriteString(", ")
			}
		}
	}
	s.WriteByte('>')
}

func (gen *Generator) writeKey(s *strings.Builder, key string, optional bool) {
	quoted := !validJSName(key)
	if quoted {
		s.WriteByte('\'')
	}
	s.WriteString(key)
	if quoted {
		s.WriteByte('\'')
	}
	if optional {
		s.WriteByte('?')
	}
	s.WriteString(": ")
}

func (gen *Generator) writeBlockType(s *strings.Builder, fields *ast.FieldList, indent int) {
	s.WriteString("{\n")

	for _, field := range fields.List {
		// if the field type is a pointer its considered to be optional
		_, optional := field.Type.(*ast.StarExpr)

		for _, fName := range field.Names {
			if !fName.IsExported() {
				continue
			}

			name := fName.Name
			gen.writeIndent(s, indent+1)
			gen.writeKey(s, name, optional)
			gen.writeExpr(s, field.Type, indent)
			s.WriteByte('\n')
		}
	}

	gen.writeIndent(s, indent)
	s.WriteByte('}')
}

func (gen *Generator) writeInterface(s *strings.Builder, node *ast.InterfaceType, indent int) {
	gen.writeBlockType(s, node.Methods, indent)
}

func (gen *Generator) writeStruct(s *strings.Builder, node *ast.StructType, indent int) {
	gen.writeBlockType(s, node.Fields, indent)
}

func (gen *Generator) writeMap(s *strings.Builder, node *ast.MapType, indent int) {
	s.WriteString("{ [key: ")
	gen.writeExpr(s, node.Key, indent)
	s.WriteString("]: ")
	gen.writeExpr(s, node.Value, indent)
	s.WriteString(" }")
}

func (gen *Generator) writeArray(s *strings.Builder, node *ast.ArrayType, indent int) {
	// cast []byte to string
	if v, ok := node.Elt.(*ast.Ident); ok && v.String() == "byte" {
		s.WriteString("string")
		return
	}

	s.WriteString("Array<")
	gen.writeExpr(s, node.Elt, indent)
	s.WriteByte('>')
}
