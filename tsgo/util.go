package tsgo

import (
	"go/ast"
	"regexp"
)

func validJsIdent(n string) bool {
	validJSNameRegexp := regexp.MustCompile(`(?m)^[\pL_][\pL\pN_]*$`)
	return validJSNameRegexp.MatchString(n)
}

func hasFieldNames(list *ast.FieldList) bool {
	return list != nil && list.NumFields() > 0 && list.List[0].Names != nil
}

// func parseIotaOffset(groupType string) int {
// 	// remove parentheses and spaces from the groupType
// 	groupType = strings.Trim(groupType, "()")
// 	if groupType == "iota" {
// 		return 0
// 	}

// 	// remove iota reference and whitespace from the type
// 	offsetStr := strings.Replace(groupType, "iota", "", 1)
// 	offsetStr = strings.ReplaceAll(offsetStr, " ", "")
	
// 	offset, err := strconv.ParseInt(offsetStr, 10, 64)
// 	if err != nil {
// 		log.Panicf("Error parsing iota offset from \"%s\": %v", offsetStr, err)
// 	}

// 	return int(offset)
// }

