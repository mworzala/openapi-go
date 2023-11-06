package util

import (
	"strings"
	"unicode"
)

func SnakeToPascalCase(str string) string {
	titleStr := strings.Title(str)
	pascalStr := strings.Replace(titleStr, "_", "", -1)
	// handle an edge case when there is a '_' at the start
	if strings.HasPrefix(str, "_") {
		pascalStr = "_" + strings.TrimLeft(pascalStr, "_")
	}

	res := []rune(pascalStr)
	for i, char := range res {
		if char == '_' {
			res[i+1] = unicode.ToUpper(res[i+1])
		}
	}
	return string(res)
}

func DashToCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool { return r == '-' })

	for index := 1; index < len(words); index++ {
		words[index] = strings.Title(words[index])
	}
	return strings.Join(words, "")
}
