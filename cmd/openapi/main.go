package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/mworzala/openapi-go/internal/pkg/gen"
	"github.com/mworzala/openapi-go/pkg/oapi"
	"gopkg.in/yaml.v3"
)

type ModelTemplate struct {
	Package string
	Schemas []*SchemaTemplate
}

type SchemaTemplate struct {
	Name   string
	Fields []*FieldTemplate
}

type FieldTemplate struct {
	Name     string
	GoName   string
	GoType   string
	Required bool
}

//go:embed model.tmpl
var modelTemplateString string

func main() {
	target := os.Args[1]
	specData, err := os.ReadFile(target)
	if err != nil {
		panic(fmt.Errorf("failed to read openapi spec: %w", err))
	}

	baseName := strings.Replace(target, ".openapi.yaml", "", -1)
	var spec oapi.Spec
	if err = yaml.Unmarshal(specData, &spec); err != nil {
		panic(fmt.Errorf("failed to unmarshal openapi spec: %w", err))
	}

	g, err := gen.New()
	if err != nil {
		panic(err)
	}

	g.GenSpecSingle(baseName, &spec)

	//baseName := strings.Replace(target, ".openapi.yaml", "", -1)
	//workDir, _ := os.Getwd()
	//packageName := path.Base(workDir)
	//
	//
	//// Execute the template into the output file
	//var context ModelTemplate
	//context.Package = packageName
	//
	//if components := spec.Components; components != nil {
	//	for _, schema := range components.Schemas {
	//		if schema.Value == nil {
	//			panic("nil schema")
	//		}
	//		if schema.Value.Ref != "" {
	//			panic("ref not supported yet")
	//		}
	//		if schema.Value.Type != "object" {
	//			panic("only objects at top level for now")
	//		}
	//
	//		var fields []*FieldTemplate
	//		for _, field := range schema.Value.Properties {
	//			fields = append(fields, &FieldTemplate{
	//				Name:     field.Name,
	//				GoName:   toPascalCase(field.Name),
	//				GoType:   "string",
	//				Required: false,
	//			})
	//		}
	//
	//		context.Schemas = append(context.Schemas, &SchemaTemplate{
	//			Name:   schema.Name,
	//			Fields: fields,
	//		})
	//	}
	//}
	//
	//tmpl, err := template.New("model.tmpl").Parse(modelTemplateString)
	//if err != nil {
	//	panic(err)
	//}
	//out := new(bytes.Buffer)
	//if err = tmpl.Execute(out, context); err != nil {
	//	panic(err)
	//}
	//
	//println(out.String())
	//
	//formatted, err := format.Source(out.Bytes())
	//if err != nil {
	//	panic(err)
	//}
	//
	//// Write the file to disk
	//modelFileName := fmt.Sprintf("model_%s.gen.go", baseName)
	//modelFile, err := os.OpenFile(modelFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	//if err != nil {
	//	panic(fmt.Errorf("failed to open file %s: %w", modelFileName, err))
	//}
	//defer modelFile.Close()
	//
	//if _, err = modelFile.Write(formatted); err != nil {
	//	panic(err)
	//}
	//
	//println("done")
}

func toPascalCase(str string) string {
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
