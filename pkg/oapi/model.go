package oapi

type Spec struct {
	Version    string             `yaml:"openapi"` // Required
	Info       *Info              `yaml:"info"`    // Required
	Paths      MapSlice[PathItem] `yaml:"paths"`
	Components *Components        `yaml:"components"`
}

//
// Info
//

type Info struct {
	Title   string `yaml:"title"`   // Required
	Version string `yaml:"version"` // Required

	BasePath string `yaml:"x-base-path"`
}

//
// API Paths
//

type PathItem struct {
	Get    *Operation `yaml:"get"`
	Put    *Operation `yaml:"put"`
	Post   *Operation `yaml:"post"`
	Delete *Operation `yaml:"delete"`
	Patch  *Operation `yaml:"patch"`
	Trace  *Operation `yaml:"trace"`
}

//
// Operations
//

type Operation struct {
	OperationId string                  `yaml:"operationId"`
	Parameters  []*ParameterOrRef       `yaml:"parameters"`
	RequestBody *RequestBody            `yaml:"requestBody"`
	Responses   MapSlice[ResponseOrRef] `yaml:"responses"`
}

type (
	ParameterOrRef struct {
		Reference `yaml:",inline"` // Only used if .Reference.Ref is set
		Parameter `yaml:",inline"`
	}
	Parameter struct {
		Name     string  `yaml:"name"` // Required
		In       string  `yaml:"in"`   // Required, todo enum of "query", "header", "path" or "cookie"
		Required bool    `yaml:"required"`
		Schema   *Schema `yaml:"schema"`
	}
)

type RequestBody struct {
	Content MapSlice[MediaType] `yaml:"content"`
}

type (
	ResponseOrRef struct {
		Reference `yaml:",inline"` // Only used if .Reference.Ref is set
		Response  `yaml:",inline"`
	}
	Response struct {
		Type    string              `yaml:"x-type"` //todo enum
		Content MapSlice[MediaType] `yaml:"content"`
	}
)

type MediaType struct {
	Schema *SchemaOrRef `yaml:"schema"`
}

//
// Components
//

type Components struct {
	Schemas    MapSlice[SchemaOrRef]    `yaml:"schemas"`
	Responses  MapSlice[ResponseOrRef]  `yaml:"responses"`
	Parameters MapSlice[ParameterOrRef] `yaml:"parameters"`
}

type (
	SchemaOrRef struct {
		Reference `yaml:",inline"` // Only used if .Reference.Ref is set
		Schema    `yaml:",inline"`
	}
	Schema struct {
		Type string `yaml:"type"` //todo enum

		// Primitive only
		Format string `yaml:"format"`

		// Objects
		Properties           MapSlice[SchemaOrRef] `yaml:"properties"`
		Required             []string              `yaml:"required"`
		AdditionalProperties bool                  `yaml:"additionalProperties"`

		// Arrays
		Items *SchemaOrRef `yaml:"items"`
	}
)

//
// Common
//

type Reference struct {
	Ref string `yaml:"$ref"`
}
