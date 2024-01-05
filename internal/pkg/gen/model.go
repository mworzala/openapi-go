package gen

type ModelTemplate struct {
	Package string
	Schemas []*TypeInfo
}

type SchemaTemplate struct {
	Name   string
	GoType string

	// Only one of the entries should be present
	Primitive *PrimitiveTemplate
	Struct    *StructTemplate
	Enum      *EnumTemplate
}

type TypeInfo struct {
	Name   string
	GoType string

	// Only one of the entries should be present
	Primitive *PrimitiveType
	Struct    *StructType
	Array     *ArrayType
	Enum      *EnumType
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

	// NEW BELOW

	PrimitiveType struct {
	}
	StructType struct {
		Fields []*FieldInfo
	}
	FieldInfo struct {
		Name string
		Type string
	}
	ArrayType struct {
		ItemGoType string
	}
	EnumType struct {
		GoType string
		Values []*EnumCase
	}
	EnumCase struct {
		Name    string
		GoValue string
	}
)
