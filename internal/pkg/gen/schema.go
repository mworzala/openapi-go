package gen

import (
	"fmt"
	"strings"

	"github.com/mworzala/openapi-go/pkg/oapi"
)

// resolve to type

func (g *Generator) resolveSchemaToType(schema *oapi.SchemaOrRef) (string, error) {
	if schema.Ref != "" {
		if existing, ok := g.schemas[schema.Ref]; ok {
			return existing.GoType, nil
		}

		// Generate the model
		schemaName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		s, ok := g.spec.Components.Schemas.Get(schemaName)
		if !ok {
			panic("failed to find schema: " + schema.Ref)
		}

		var err error
		generated, err := g.genSingleSchema(schemaName, s)
		if err != nil {
			panic(err)
		}
		g.schemas[schema.Ref] = generated

		return generated.GoType, nil
	}

	switch schema.Type {
	case "string":
		goType := "string"
		switch schema.Format {
		case "date-time":
			goType = "time.Time"
		case "binary":
			goType = "[]byte"
		default:
			if schema.Format != "" {
				println("unsupported string format: " + schema.Format)
			}
		}
		return goType, nil
	default:
		panic("unsupported schema type: " + schema.Type)
	}

	//todo
	panic("")
}

func (g *Generator) genSingleSchema(shortName string, schema *oapi.SchemaOrRef) (*SchemaTemplate, error) {
	var result SchemaTemplate
	result.Name = shortName

	if schema.Ref != "" {
		panic("refs not supported")
	}

	switch schema.Type {
	case "string":
		goType := "string"
		switch schema.Format {
		case "date-time":
			goType = "time.Time"
		case "binary":
			goType = "[]byte"
		default:
			if schema.Format != "" {
				println("unsupported string format: " + schema.Format)
			}
		}

		result.Name = goType
		result.Primitive = &PrimitiveTemplate{
			Name: goType,
			Type: goType,
		}
	case "object":
		// Special case: if object with no fields and additionalProperties=true, use a Go `any` type.
		if len(schema.Properties) == 0 && schema.AdditionalProperties {
			result.Name = "interface{}"
			result.Primitive = &PrimitiveTemplate{
				Name: "interface{}",
				Type: "interface{}",
			}
			break
		}

		var fields []*FieldTemplate
		for _, field := range schema.Properties {
			ty, err := g.genSingleSchema("", field.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to generate field %s.%s: %w", shortName, field.Name, err)
			}

			fields = append(fields, &FieldTemplate{
				Name: field.Name,
				Type: ty.Name,
			})
		}

		result.Struct = &StructTemplate{
			Name:   shortName,
			Fields: fields,
		}
		result.GoType = "*" + result.Struct.Name
	default:
		panic("unsupported schema type: " + schema.Type)
	}

	if result.GoType == "" {
		result.GoType = result.Name
	}
	return &result, nil
}

func (g *Generator) oapiTypeToGoType(oapiType string) string {
	panic("")
}
