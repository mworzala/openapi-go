package util

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

func CamelToPascalCase(str string) string {
	// Uppercase first character, leave the rest untouched
	return cases.Title(language.English, cases.NoLower).String(str)
}

func SnakeToPascalCase(str string) string {
	titleStr := titleSnakeCase(str)
	pascalStr := strings.Replace(titleStr, "_", "", -1)
	return pascalStr
}

func titleSnakeCase(str string) string {
	components := strings.Split(str, "_")
	for i, component := range components {
		components[i] = cases.Title(language.English, cases.NoLower).String(component)
	}
	return strings.Join(components, "_")
}

func DashToCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool { return r == '-' })

	for index := 1; index < len(words); index++ {
		words[index] = strings.Title(words[index])
	}
	return strings.Join(words, "")
}
