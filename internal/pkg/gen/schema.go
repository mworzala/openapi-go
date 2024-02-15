package gen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mworzala/openapi-go/internal/pkg/util"
	"github.com/mworzala/openapi-go/pkg/oapi"
)

// resolve to type

func (g *Generator) resolveSchemaToType(schema *oapi.AnySchema) (string, error) {
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
	case "number":
		goType := "float64"
		return goType, nil
	case "array":
		ty, err := g.resolveSchemaToType(schema.Items)
		if err != nil {
			return "", fmt.Errorf("failed to resolve array type: %w", err)
		}

		return "[]" + ty, nil
	default:
		panic("unsupported schema type 2: " + schema.Type)
	}

	//todo
	panic("")
}

func (g *Generator) genSingleSchema(shortName string, schema *oapi.AnySchema) (*SchemaTemplate, error) {
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
	case "number":
		goType := "float64"
		result.Name = goType
		result.Primitive = &PrimitiveTemplate{
			Name: goType,
			Type: goType,
		}
	case "object":
		// Special case: if object with no fields and additionalProperties=true, use a `map[string]interface{}` type.
		if len(schema.Properties) == 0 && schema.AdditionalProperties {
			if schema.Name != "" {
				result.TypeAlias = &TypeAliasTemplate{
					Name: shortName,
					Type: "map[string]interface{}",
				}
				result.GoType = "*" + result.TypeAlias.Name
			} else {
				result.Name = "map[string]interface{}"
				result.Primitive = &PrimitiveTemplate{
					Name: "map[string]interface{}",
					Type: "map[string]interface{}",
				}
			}
			break
		}

		var fields []*FieldTemplate
		for _, field := range schema.Properties {
			ty, err := g.resolveSchemaToType(field.Value)
			if err != nil {
				return nil, fmt.Errorf("failed to generate field %s.%s: %w", shortName, field.Name, err)
			}

			fields = append(fields, &FieldTemplate{
				Name: field.Name,
				Type: ty,
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

// Newer stuff

func (g *Generator) resolveSchema(schema *oapi.AnySchema, name string, anonymous, required bool) (*TypeInfo, error) {
	if schema.Ref != "" {
		if existing, ok := g.schemas2.Get(schema.Ref); ok {
			return existing, nil
		}

		// Generate the model
		schemaName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		s, ok := g.spec.Components.Schemas.Get(schemaName)
		if !ok {
			return nil, fmt.Errorf("missing reference: %s", schema.Ref)
		}

		//todo this should be valid and create an alias type
		if s.Ref != "" {
			return nil, fmt.Errorf("double reference: %s -> %s", schema.Ref, s.Ref)
		}

		generated, err := g.generateTypeFromSchemaNoRef(s, schemaName, required)
		if err != nil {
			return nil, fmt.Errorf("gen fail for %s: %w", schema.Ref, err)
		}

		// This is a named type, so we need to 'rename' it.
		//todo how do i do
		//generated.Name = schemaName

		g.schemas2 = g.schemas2.With(schema.Ref, generated)
		return generated, nil
	}

	ti, err := g.generateTypeFromSchemaNoRef(schema, name, required)
	if err != nil {
		return nil, fmt.Errorf("gen fail for %s: %w", schema.Ref, err)
	}
	// If it is a named type we should store it
	if !anonymous || ti.Enum != nil {
		g.schemas2 = g.schemas2.With(fmt.Sprintf("#/components/schemas/%s", name), ti)
	}
	return ti, nil
}

func (g *Generator) generateTypeFromSchemaNoRef(schema *oapi.AnySchema, nameOverride string, required bool) (*TypeInfo, error) {

	if len(schema.AllOf) > 0 {
		return g.generateAllOfType(schema.AllOf, nameOverride, required)
	}

	// Assume it is a normal type
	if schema.Type == "" {
		return nil, fmt.Errorf("missing schema type")
	}
	switch schema.Type {
	case "string":
		return g.generateStringType(&schema.Schema, nameOverride, required)
	case "number", "integer":
		return g.generateNumericType(&schema.Schema, nameOverride, required)
	case "boolean":
		return g.generateBooleanType(&schema.Schema, required)
	case "object":
		return g.generateObjectType(&schema.Schema, nameOverride, required)
	case "array":
		return g.generateArrayType(&schema.Schema, nameOverride, required)
	default:
		return nil, fmt.Errorf("unsupported schema type: %s", schema.Type)
	}
}

func (g *Generator) generateStringType(schema *oapi.Schema, nameOverride string, required bool) (*TypeInfo, error) {
	var result TypeInfo
	result.Name = "string"
	result.ZeroValue = "\"\""
	result.Primitive = &PrimitiveType{}

	switch schema.Format {
	case "date-time", "date":
		result.GoType = "time.Time"
	case "byte":
		panic("todo should be base64 encoded")
	case "binary":
		result.GoType = "[]byte"
		result.ZeroValue = "nil"
	case "uuid":
		if !slices.Contains(g.extraImports, "github.com/google/uuid") {
			g.extraImports = append(g.extraImports, "github.com/google/uuid")
		}
		result.GoType = "uuid.UUID"
	default:
		if schema.Format != "" {
			println(fmt.Sprintf("unsupported string format: %s", schema.Format))
		}
		result.GoType = "string"
		result.ZeroValue = "\"\""
	}

	r, err := g.maybeGenerateEnum(schema, &result, nameOverride, required)
	if err != nil {
		return nil, err
	}
	if !required {
		r.GoType = "*" + r.GoType
		r.ZeroValue = "nil"
	}
	return r, nil
}

func (g *Generator) generateNumericType(schema *oapi.Schema, nameOverride string, required bool) (*TypeInfo, error) {
	var result TypeInfo
	result.Primitive = &PrimitiveType{}
	result.ZeroValue = "0"

	// Right now this is a bit too expressive. `integer` and `number` are treated (almost) identically so you
	// can have an integer with a float format. This should be fixed later.

	// Any integer type, will check floats later if this is allowed to be a float
	switch schema.Format {
	// Valid to be any go int type and will use that directly
	case "int8", "int16", "int32", "int64", "int", "uint8", "uint16", "uint32", "uint64", "uint":
		result.GoType = schema.Format
	case "float32", "float64":
		result.GoType = schema.Format
	default:
		if schema.Format != "" {
			println(fmt.Sprintf("unsupported numeric format: %s", schema.Format))
		}

		if schema.Type == "integer" {
			result.GoType = "int"
		} else {
			result.GoType = "float64"
		}
	}

	result.Name = result.GoType

	r, err := g.maybeGenerateEnum(schema, &result, nameOverride, required)
	if err != nil {
		return nil, err
	}
	if !required {
		r.GoType = "*" + r.GoType
	}
	return r, nil
}

func (g *Generator) maybeGenerateEnum(schema *oapi.Schema, parent *TypeInfo, name string, required bool) (*TypeInfo, error) {
	if len(schema.Enum) == 0 {
		return parent, nil
	}

	var result TypeInfo
	if schema.Name != "" {
		result.Name = schema.Name
		result.GoType = schema.Name
	} else {
		result.Name = name
		result.GoType = name
	}
	result.ZeroValue = fmt.Sprintf("%s(%s)", result.GoType, parent.ZeroValue)

	result.Enum = &EnumType{}
	result.Enum.GoType = parent.GoType
	for i, value := range schema.Enum {
		entry := &EnumCase{
			Name: value,
		}
		if result.Enum.GoType == "string" {
			entry.GoValue = fmt.Sprintf("\"%s\"", value)
		} else {
			entry.GoValue = fmt.Sprintf("%d", i)
		}
		result.Enum.Values = append(result.Enum.Values, entry)
	}

	if !required {
		result.GoType = "*" + result.GoType
		result.ZeroValue = "nil"
	}

	return &result, nil
}

func (g *Generator) generateBooleanType(schema *oapi.Schema, required bool) (*TypeInfo, error) {
	result := &TypeInfo{
		Name:      "boolean",
		GoType:    "bool",
		ZeroValue: "false",
		Primitive: &PrimitiveType{},
	}

	if !required {
		result.GoType = "*bool"
		result.ZeroValue = "nil"
	}

	return result, nil
}

func (g *Generator) generateObjectType(schema *oapi.Schema, nameOverride string, required bool) (*TypeInfo, error) {
	var result TypeInfo
	if schema.Name != "" {
		result.Name = schema.Name
		result.GoType = schema.Name
	} else if nameOverride != "" {
		result.Name = nameOverride
		result.GoType = nameOverride
	}
	result.ZeroValue = "nil"

	if len(schema.Properties) == 0 && schema.AdditionalProperties {
		if schema.Name != "" {
			result.TypeAlias = &TypeAliasType{
				AliasGoType: "map[string]interface{}",
			}
		} else {
			result.GoType = "map[string]interface{}"
		}
		return &result, nil
	}

	result.GoType = fmt.Sprintf("*%s", result.GoType)
	result.Struct = &StructType{}

	for _, field := range schema.Properties {
		fieldRequired := slices.Contains(schema.Required, field.Name)

		fieldType, err := g.resolveSchema(field.Value, result.Name+util.CamelToPascalCase(field.Name),
			field.Value.Type != "object", fieldRequired)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve field type: %w", err)
		}

		fieldGoType := fieldType.GoType
		if fieldType.Enum != nil && !fieldRequired {
			fieldGoType = "*" + fieldGoType
		}

		result.Struct.Fields = append(result.Struct.Fields, &FieldInfo{
			Name: field.Name,
			Type: fieldGoType,
		})
	}

	return &result, nil
}

func (g *Generator) generateArrayType(schema *oapi.Schema, name string, required bool) (*TypeInfo, error) {
	var result TypeInfo
	result.Name = name
	result.ZeroValue = "nil"
	result.Array = &ArrayType{}

	itemType, err := g.resolveSchema(schema.Items, name+"Item", false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve array type: %w", err)
	}
	result.Array.ItemGoType = itemType.GoType

	if schema.Name != "" {
		result.GoType = schema.Name
	} else {
		result.GoType = fmt.Sprintf("[]%s", itemType.GoType)
	}

	//if itemType.Primitive != nil {
	//	result.Name = name
	//	result.GoType = fmt.Sprintf("[]%s", name)
	//}

	return &result, nil
}

func (g *Generator) generateAllOfType(schema []*oapi.AnySchema, name string, required bool) (*TypeInfo, error) {
	var result TypeInfo
	result.Name = "allOf"
	result.ZeroValue = "nil"
	result.Struct = &StructType{}

	for _, subSchema := range schema {
		subType, err := g.resolveSchema(subSchema, "anonymous", true, required)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve allOf type: %w", err)
		}

		// If it is anonymous, append all the fields. Otherwise add as an embedded struct
		if subType.Name == "anonymous" {
			for _, field := range subType.Struct.Fields {
				result.Struct.Fields = append(result.Struct.Fields, field)
			}
		} else {
			result.Struct.Fields = append(result.Struct.Fields, &FieldInfo{
				Name: "", // Will embed it
				Type: subType.GoType,
			})
		}
	}

	if name != "" {
		result.Name = name
		result.GoType = fmt.Sprintf("*%s", name)
	}

	return &result, nil
}
