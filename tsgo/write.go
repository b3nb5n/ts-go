package tsgo

import (
	"fmt"
	"strings"

	"go/ast"
	"go/token"

	"github.com/fatih/structtag"
)

func (gen *Generator) writeByte(b byte) {
	gen.out = append(gen.out, b)
}

func (gen *Generator) writeString(s string) {
	gen.out = append(gen.out, s...)
}

func (gen *Generator) writeIndent(depth int) {
	gen.writeString(strings.Repeat(gen.config.Indent, depth))
}

func (gen *Generator) writeType(t ast.Expr, depth int, optionalParens bool) {
	switch t := t.(type) {
	case *ast.StarExpr:
		if optionalParens {
			gen.writeByte('(')
		}
		gen.writeType(t.X, depth, false)
		gen.writeString(" | undefined")
		if optionalParens {
			gen.writeByte(')')
		}
	case *ast.ArrayType:
		if v, ok := t.Elt.(*ast.Ident); ok && v.String() == "byte" {
			gen.writeString("string")
			break
		}
		gen.writeType(t.Elt, depth, true)
		gen.writeString("[]")
	case *ast.StructType:
		gen.writeString("{\n")
		gen.writeStructFields(t.Fields.List, depth+1)
		gen.writeIndent(depth + 1)
		gen.writeByte('}')
	case *ast.Ident:
		gen.writeString(getIdent(t.String()))
	case *ast.SelectorExpr:
		// e.g. `time.Time`
		longType := fmt.Sprintf("%s.%s", t.X, t.Sel)
		mappedTsType, ok := gen.config.TypeMappings[longType]
		if ok {
			gen.writeString(mappedTsType)
		} else { // For unknown types we put `any`
			gen.writeString("any")
			gen.writeString(" /* ")
			gen.writeString(longType)
			gen.writeString(" */")
		}
	case *ast.MapType:
		gen.writeString("{ [key: ")
		gen.writeType(t.Key, depth, false)
		gen.writeString("]: ")
		gen.writeType(t.Value, depth, false)
		gen.writeByte('}')
	case *ast.BasicLit:
		gen.writeString(t.Value)
	case *ast.ParenExpr:
		gen.writeByte('(')
		gen.writeType(t.X, depth, false)
		gen.writeByte(')')
	case *ast.BinaryExpr:
		gen.writeType(t.X, depth, false)
		gen.writeByte(' ')
		gen.writeString(t.Op.String())
		gen.writeByte(' ')
		gen.writeType(t.Y, depth, false)
	case *ast.InterfaceType:
		gen.writeInterfaceFields(t.Methods.List, depth+1)
	case *ast.CallExpr, *ast.FuncType, *ast.ChanType:
		gen.writeString("any")
	case *ast.UnaryExpr:
		if t.Op == token.TILDE {
			// We just ignore the tilde token, in Typescript extended types are
			// put into the generic typing itself, which we can't support yet.
			gen.writeType(t.X, depth, false)
		} else {
			err := fmt.Errorf("unhandled unary expr: %v\n %T", t, t)
			fmt.Println(err)
			panic(err)
		}
	case *ast.IndexListExpr:
		gen.writeType(t.X, depth, false)
		gen.writeByte('<')
		for i, index := range t.Indices {
			gen.writeType(index, depth, false)
			if i != len(t.Indices)-1 {
				gen.writeString(", ")
			}
		}
		gen.writeByte('>')
	case *ast.IndexExpr:
		gen.writeType(t.X, depth, false)
		gen.writeByte('<')
		gen.writeType(t.Index, depth, false)
		gen.writeByte('>')
	default:
		err := fmt.Errorf("unhandled: %s\n %T", t, t)
		fmt.Println(err)
		panic(err)
	}
}

func (gen *Generator) writeTypeParamsFields(fields []*ast.Field) {
	gen.writeByte('<')
	for i, f := range fields {
		for j, ident := range f.Names {
			gen.writeString(ident.Name)
			gen.writeString(" extends ")
			gen.writeType(f.Type, 0, true)

			if i != len(fields)-1 || j != len(f.Names)-1 {
				gen.writeString(", ")
			}
		}
	}
	gen.writeByte('>')
}

func (gen *Generator) writeInterfaceFields(fields []*ast.Field, depth int) {
	if len(fields) == 0 { // Type without any fields (probably only has methods)
		gen.writeString("any")
		return
	}
	gen.writeByte('\n')
	for _, f := range fields {
		if _, isFunc := f.Type.(*ast.FuncType); isFunc {
			continue
		}
		gen.writeCommentGroupIfNotNil(f.Doc, depth+1)
		gen.writeIndent(depth + 1)
		gen.writeType(f.Type, depth, false)

		if f.Comment != nil {
			gen.writeString(" // ")
			gen.writeString(f.Comment.Text())
		}
	}
}

func (gen *Generator) writeStructFields(fields []*ast.Field, depth int) {
	for _, f := range fields {
		optional := false
		required := false

		var fieldName string
		if len(f.Names) != 0 && f.Names[0] != nil && len(f.Names[0].Name) != 0 {
			fieldName = f.Names[0].Name
		}
		if len(fieldName) == 0 || 'A' > fieldName[0] || fieldName[0] > 'Z' {
			continue
		}

		var name string
		var tstype string
		if f.Tag != nil {
			tags, err := structtag.Parse(f.Tag.Value[1 : len(f.Tag.Value)-1])
			if err != nil {
				panic(err)
			}

			jsonTag, err := tags.Get("json")
			if err == nil {
				name = jsonTag.Name
				if name == "-" {
					continue
				}

				optional = jsonTag.HasOption("omitempty")
			}
			tstypeTag, err := tags.Get("tstype")
			if err == nil {
				tstype = tstypeTag.Name
				if tstype == "-" {
					continue
				}
				required = tstypeTag.HasOption("required")
			}
		}

		if len(name) == 0 {
			name = fieldName
		}

		gen.writeCommentGroupIfNotNil(f.Doc, depth+1)

		gen.writeIndent(depth + 1)
		quoted := !validJSName(name)
		if quoted {
			gen.writeByte('\'')
		}
		gen.writeString(name)
		if quoted {
			gen.writeByte('\'')
		}

		switch t := f.Type.(type) {
		case *ast.StarExpr:
			optional = !required
			f.Type = t.X
		}

		if optional {
			gen.writeByte('?')
		}

		gen.writeString(": ")

		if tstype == "" {
			gen.writeType(f.Type, depth, false)
		} else {
			gen.writeString(tstype)
		}
		gen.writeByte(';')

		if f.Comment != nil {
			// Line comment is present, that means a comment after the field.
			gen.writeString(" // ")
			gen.writeString(f.Comment.Text())
		} else {
			gen.writeByte('\n')
		}
	}
}
