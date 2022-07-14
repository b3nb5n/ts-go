package tsgo

import (
	"fmt"
	"go/ast"
	"regexp"
	"strconv"
	"strings"
)

func validJsIdent(n string) bool {
	validJSNameRegexp := regexp.MustCompile(`(?m)^[\pL_][\pL\pN_]*$`)
	return validJSNameRegexp.MatchString(n)
}

func hasFieldNames(list *ast.FieldList) bool {
	return list != nil && list.NumFields() > 0 && list.List[0].Names != nil
}

func valueReferencesIota(valStr string) bool {
	return strings.Contains(valStr, "iota")
}

func parseIotaOffset(valStr string) (int, error) {
	// remove parentheses and whitespace from the groupType
	valStr = strings.Trim(valStr, "()")
	valStr = strings.ReplaceAll(valStr, " ", "")
	if valStr == "iota" {
		return 0, nil
	}

	// remove iota reference from the type
	offsetStr := strings.Replace(valStr, "iota", "", 1)
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsing iota offset from \"%s\": %v", offsetStr, err)
	}

	return int(offset), nil
}

