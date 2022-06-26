package tsgo

import (
	"fmt"
	"go/ast"
	"strings"
)

type groupContext struct {
	isGroupedDeclaration bool
	doc                  *ast.CommentGroup
	groupValue           string
	groupType            string
	iotaValue            int
	iotaOffset           int
}

func (gen *Generator) writeDecl(decl *ast.GenDecl) {
	// This checks whether the declaration is a group declaration like:
	// const (
	// 	  X = 3
	//    Y = "abc"
	// )
	isGroupedDeclaration := len(decl.Specs) > 1
	if !isGroupedDeclaration {
		gen.writeCommentGroupIfNotNil(decl.Doc, 0)
	}

	// We need a bit of state to handle syntax like
	// const (
	//   X SomeType = iota
	//   _
	//   Y
	//   Foo string = "Foo"
	//   _
	//   AlsoFoo
	// )
	group := &groupContext{
		isGroupedDeclaration: isGroupedDeclaration,
		doc:                  decl.Doc,
		groupType:            "",
		groupValue:           "",
		iotaValue:            -1,
	}

	for _, spec := range decl.Specs {
		gen.writeSpec(spec, group)
	}
}

func (gen *Generator) writeSpec(spec ast.Spec, group *groupContext) {
	// e. "type Foo struct {}" or "type Bar = string"
	ts, ok := spec.(*ast.TypeSpec)
	if ok && ts.Name.IsExported() {
		gen.writeTypeSpec(ts, group)
	}

	// e. "const Foo = 123"
	vs, ok := spec.(*ast.ValueSpec)
	if ok {
		gen.writeValueSpec(vs, group)
	}
}

// Writing of type specs, which are expressions like
// `type X struct { ... }`
// or
// `type Bar = string`
func (gen *Generator) writeTypeSpec(ts *ast.TypeSpec, group *groupContext) {
	if ts.Doc != nil { // The spec has its own comment, which overrules the grouped comment.
		gen.writeCommentGroup(ts.Doc, 0)
	} else if group.isGroupedDeclaration {
		gen.writeCommentGroupIfNotNil(group.doc, 0)
	}

	st, isStruct := ts.Type.(*ast.StructType)
	if isStruct {
		gen.writeString("export interface ")
		gen.writeString(ts.Name.Name)
		if ts.TypeParams != nil {
			gen.writeTypeParamsFields(ts.TypeParams.List)
		}

		gen.writeString(" {\n")
		gen.writeStructFields(st.Fields.List, 0)
		gen.writeString("}")
	}

	id, isIdent := ts.Type.(*ast.Ident)
	if isIdent {
		gen.writeString("export type ")
		gen.writeString(ts.Name.Name)
		gen.writeString(" = ")
		gen.writeString(getIdent(id.Name))
		gen.writeString(";")
	}

	if !isStruct && !isIdent {
		gen.writeString("export type ")
		gen.writeString(ts.Name.Name)
		gen.writeString(" = ")
		gen.writeType(ts.Type, 0, true)
		gen.writeString(";")
	}

	if ts.Comment != nil {
		gen.writeString(" // " + ts.Comment.Text())
	} else {
		gen.writeString("\n")
	}
}

// Writign of value specs, which are exported const expressions like
// const SomeValue = 3
func (gen *Generator) writeValueSpec(vs *ast.ValueSpec, group *groupContext) {
	for i, name := range vs.Names {
		group.iotaValue = group.iotaValue + 1
		if name.Name == "_" || !name.IsExported() {
			continue
		}

		if vs.Doc != nil { // The spec has its own comment, which overrules the grouped comment.
			gen.writeCommentGroup(vs.Doc, 0)
		} else if group.isGroupedDeclaration {
			gen.writeCommentGroupIfNotNil(group.doc, 0)
		}

		hasExplicitValue := len(vs.Values) > i
		if hasExplicitValue {
			group.groupType = ""
		}

		gen.writeString("export const ")
		gen.writeString(name.Name)
		if vs.Type != nil {
			gen.writeString(": ")

			tempSB := &strings.Builder{}
			gen.writeType(vs.Type, 0, true)
			typeString := tempSB.String()

			gen.writeString(typeString)
			group.groupType = typeString
		} else if group.groupType != "" && !hasExplicitValue {
			gen.writeString(": ")
			gen.writeString(group.groupType)
		}

		gen.writeString(" = ")

		if hasExplicitValue {
			val := vs.Values[i]

			tmpGen := NewGenerator(gen.config)
			tmpGen.writeType(val, 0, true)
			valueString := tmpGen.String()

			if isProbablyIotaType(valueString) {
				group.iotaOffset = basicIotaOffsetValueParse(valueString)
				group.groupValue = "iota"
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			} else {
				group.groupValue = valueString
			}
			gen.writeString(valueString)

		} else { // We must use the previous value or +1 in case of iota
			valueString := group.groupValue
			if group.groupValue == "iota" {
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			}
			gen.writeString(valueString)
		}

		gen.writeByte(';')
		if vs.Comment != nil {
			gen.writeString(" // " + vs.Comment.Text())
		} else {
			gen.writeByte('\n')
		}
	}
}
