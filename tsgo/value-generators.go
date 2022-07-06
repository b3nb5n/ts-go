package tsgo

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func writeConstantValue(s *strings.Builder, expr ast.Expr) error {
	switch expr := expr.(type) {
	case *ast.Ident:
		s.WriteString(expr.Name)
	case *ast.BasicLit:
		if expr.Kind == token.IMAG {
			return fmt.Errorf("No ts equivilent for imaginary numbers")
		} else if expr.Kind == token.CHAR {
			s.WriteString("\"" + expr.Value[1:len(expr.Value) - 1] + "\"")
		} else {
			s.WriteString(expr.Value)
		}
	case *ast.BinaryExpr:
		writeConstantValue(s, expr.X)
		s.WriteString(" " + expr.Op.String() + " ")
		writeConstantValue(s, expr.Y)
	default:
		return fmt.Errorf("Unrecognized constant value")
	}

	return nil
}