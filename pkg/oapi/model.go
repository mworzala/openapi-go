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

	BasePath *string `yaml:"x-base-path"`
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

func (p PathItem) IsEmpty() bool {
	return p.Get == nil && p.Put == nil && p.Post == nil &&
		p.Delete == nil && p.Patch == nil && p.Trace == nil
}

//
// Operations
//

type Operation struct {
	OperationId string                  `yaml:"operationId"`
	Parameters  []*ParameterOrRef       `yaml:"parameters"`
	RequestBody *RequestBodyOrRef       `yaml:"requestBody"`
	Responses   MapSlice[ResponseOrRef] `yaml:"responses"`
}

type (
	ParameterOrRef struct {
		Reference `yaml:",inline"` // Only used if .Reference.Ref is set
		Parameter `yaml:",inline"`
	}
	Parameter struct {
		Name       string  `yaml:"name"` // Required
		CustomName string  `yaml:"x-name"`
		In         string  `yaml:"in"` // Required, todo enum of "query", "header", "path" or "cookie"
		Required   bool    `yaml:"required"`
		Schema     *Schema `yaml:"schema"`

		// Query only
		Explode bool `yaml:"explode"`
	}
)

type (
	RequestBodyOrRef struct {
		Reference   `yaml:",inline"` // Only used if .Reference.Ref is set
		RequestBody `yaml:",inline"`
	}
	RequestBody struct {
		Content MapSlice[MediaType] `yaml:"content"`
	}
)

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
	Schema *AnySchema `yaml:"schema"`
}

//
// Components
//

type Components struct {
	Schemas       MapSlice[AnySchema]        `yaml:"schemas"`
	Parameters    MapSlice[ParameterOrRef]   `yaml:"parameters"`
	RequestBodies MapSlice[RequestBodyOrRef] `yaml:"requestBodies"`
	Responses     MapSlice[ResponseOrRef]    `yaml:"responses"`
}

type (
	AnySchema struct {
		Reference `yaml:",inline"` // Only used if .Reference.Ref is set
		Schema    `yaml:",inline"`
		AllOf     []*AnySchema `yaml:"allOf"`
	}
	Schema struct {
		Type string `yaml:"type"` //todo enum
		Name string `yaml:"x-name"`

		// Primitive only
		Format string   `yaml:"format"`
		Enum   []string `yaml:"enum"`

		// Objects
		Required             []string            `yaml:"required"`
		Properties           MapSlice[AnySchema] `yaml:"properties"`
		AdditionalProperties bool                `yaml:"additionalProperties"`

		// Arrays
		Items *AnySchema `yaml:"items"`
	}
)

//
// Common
//

type Reference struct {
	Ref string `yaml:"$ref"`
}
