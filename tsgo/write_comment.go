package tsgo

import (
	"go/ast"
	"strings"
)

func (gen *Generator) writeCommentGroupIfNotNil(cg *ast.CommentGroup, depth int) {
	if cg != nil {
		gen.writeCommentGroup(cg, depth)
	}
}

func (gen *Generator) writeCommentGroup(cg *ast.CommentGroup, depth int) {
	docLines := strings.Split(cg.Text(), "\n")

	gen.writeIndent(depth)
	gen.writeString("/**\n")

	for _, c := range docLines {
		if len(strings.TrimSpace(c)) == 0 {
			continue
		}
		gen.writeIndent(depth)
		gen.writeString(" * ")
		c = strings.ReplaceAll(c, "*/", "*\\/") // An edge case: a // comment can contain */
		gen.writeString(c)
		gen.writeByte('\n')
	}

	gen.writeIndent(depth)
	gen.writeString(" */\n")
}
