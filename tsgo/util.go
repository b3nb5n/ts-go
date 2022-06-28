package tsgo

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var validJSNameRegexp = regexp.MustCompile(`(?m)^[\pL_][\pL\pN_]*$`)

func validJSName(n string) bool {
	return validJSNameRegexp.MatchString(n)
}

func getIdent(s string) string {
	switch s {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"complex64", "complex128":
		return "number"
	case "bool":
		return "boolean"
	case "byte", "rune", "string":
		return "string"
	}

	return s
}

func parseIotaOffset(groupType string) int {
	// remove parentheses and spaces from the groupType
	groupType = strings.Trim(groupType, "()")
	if groupType == "iota" {
		return 0
	}

	// remove iota reference and whitespace from the type
	offsetStr := strings.Replace(groupType, "iota", "", 1)
	offsetStr = strings.ReplaceAll(offsetStr, " ", "")
	
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		log.Panicf("Error parsing iota offset from \"%s\": %v", offsetStr, err)
	}

	return int(offset)
}

