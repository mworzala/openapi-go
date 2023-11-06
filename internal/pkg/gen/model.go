package gen

type ModelTemplate struct {
	Package string
	Schemas []*SchemaTemplate
}

type SchemaTemplate struct {
	Name   string
	GoType string

	// Only one of the entries should be present
	Primitive *PrimitiveTemplate
	Struct    *StructTemplate
	Enum      *EnumTemplate
}

type (
	PrimitiveTemplate struct {
		Name string
		Type string
	}
	StructTemplate struct {
		Name   string
		Fields []*FieldTemplate
	}
	FieldTemplate struct {
		Name string
		Type string
	}
	EnumTemplate struct {
		Name   string
		Type   string // string, int, etc
		Values []string
	}
)
