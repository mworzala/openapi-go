package gen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mworzala/openapi-go/internal/pkg/util"
	"github.com/mworzala/openapi-go/pkg/oapi"
)

type ServerTemplate struct {
	Package string
	Name    string

	BasePath   string
	Operations []*OperationTemplate

	UseFx bool
}

type OperationTemplate struct {
	Name   string
	Method string
	Path   string

	// Request Params
	PathParams   []*ParamTemplate
	QueryParams  []*QueryParamTemplate
	HeaderParams []*ParamTemplate
	Body         *RequestBodyTemplate

	// Response
	Response *ResponseTemplate
}

type RequestBodyTemplate struct {
	GoType string
	IsRaw  bool
}

type QueryParamTemplate struct {
	Name     string
	Required bool

	StructGoType string
}

type ParamTemplate struct {
	Name       string
	CustomName string
	Required   bool
}

type ResponseTemplate struct {
	Cases []*ResponseCaseTemplate

	// One of the following
	Single *SingleResponseTemplate

	//todo there are three cases
	// - single code, single content type
	// - single code, multiple content types
	// - multiple codes, multiple content types

	EmptyCode *int // If the response is empty, this is the code to use
}

type ResponseCaseTemplate struct {
	Name   string
	GoType string

	Single *SingleResponseTemplate
	Multi  *[]*SingleResponseTemplate
}

type SingleResponseTemplate struct {
	Code        int
	ContentType string
	Name        string
}

func (g *Generator) genOperation(path, method string, op *oapi.Operation) (*OperationTemplate, error) {
	result := OperationTemplate{Method: method, Path: path}

	if op.OperationId == "" {
		return nil, fmt.Errorf("'%s'.%s: missing operationId", path, method)
	}
	result.Name = op.OperationId

	for _, param := range op.Parameters {
		switch param.In {
		case "path":
			result.PathParams = append(result.PathParams, &ParamTemplate{Name: param.Name})
		case "query":
			tmpl := &QueryParamTemplate{
				Name:     param.Name,
				Required: param.Required,
			}
			result.QueryParams = append(result.QueryParams, tmpl)

			if param.Schema.Type == "string" {
				// Already handled by above
			} else if param.Schema.Type == "object" {
				if !param.Explode {
					return nil, fmt.Errorf("'%s'.%s: query param '%s' is an object, but explode is not set", path, method, param.Name)
				}

				typeName := fmt.Sprintf("%s%s", util.CamelToPascalCase(result.Name), util.CamelToPascalCase(param.Name))
				typeInfo, err := g.resolveSchema(&oapi.AnySchema{Schema: *param.Schema}, typeName, false)
				if err != nil {
					return nil, fmt.Errorf("'%s'.%s: failed to resolve query param '%s': %w", path, method, param.Name, err)
				}

				tmpl.StructGoType = typeInfo.GoType
			} else {
				panic("query params must be string or object")
			}
		case "header":
			var paramName string
			if param.CustomName != "" {
				paramName = param.CustomName
			} else {
				paramName = param.Name
			}

			result.HeaderParams = append(result.HeaderParams, &ParamTemplate{
				Name:       param.Name,
				CustomName: paramName,
				Required:   param.Required,
			})
		default:
			println("unsupported param type: " + param.In)
		}
	}

	if op.RequestBody != nil {
		var err error
		result.Body, err = g.genSingleRequestBody(result.Name, op.RequestBody)
		if err != nil {
			return nil, fmt.Errorf("'%s'.%s: failed to generate request body: %w", path, method, err)
		}
	}

	typedResponses := make(oapi.MapSlice[oapi.ResponseOrRef], 0)
	for _, response := range op.Responses {
		if response.Value.Ref != "" || response.Value.Type == "empty" {
			typedResponses = typedResponses.With(response.Name, response.Value)
			continue
		}

		for _, content := range response.Value.Content {
			if content.Value.Schema != nil {
				typedResponses = typedResponses.With(response.Name, response.Value)
				break
			}
		}
	}

	var res ResponseTemplate
	if len(typedResponses) > 0 {
		for _, response := range typedResponses {
			code, err := strconv.Atoi(response.Name)
			if err != nil {
				return nil, fmt.Errorf("'%s'.%s: failed to parse response code '%s': %w", path, method, response.Name, err)
			}

			if response.Value.Ref == "" && response.Value.Type == "empty" {
				res.EmptyCode = &code

				// If there is no response body, don't generate anything and keep going.
				if len(response.Value.Content) == 0 {
					continue
				}
			}

			baseName := fmt.Sprintf("%sResponse", util.CamelToPascalCase(result.Name))
			resCase, err := g.genSingleResponse(baseName, code, response.Value)
			if err != nil {
				return nil, fmt.Errorf("'%s'.%s: failed to generate response: %w", path, method, err)
			}

			res.Cases = append(res.Cases, resCase)
		}
	}
	result.Response = &res

	return &result, nil
}

func (g *Generator) genSingleRequestBody(baseName string, model *oapi.RequestBodyOrRef) (*RequestBodyTemplate, error) {
	if model.Ref != "" {
		path := strings.Replace(model.Ref, "#/components/requestBodies/", "", 1)
		ref, ok := g.spec.Components.RequestBodies.Get(path)
		if !ok {
			return nil, fmt.Errorf("failed to find request body ref: %s", model.Ref)
		}

		return g.genSingleRequestBody(path, ref)
	}

	for _, content := range model.Content {
		bodyStructName := fmt.Sprintf("%sRequest", util.CamelToPascalCase(baseName))
		if content.Name == "application/json" {
			s, err := g.resolveSchema(content.Value.Schema, bodyStructName, false)
			if err != nil {
				panic(err)
			}
			template := &RequestBodyTemplate{GoType: s.GoType}
			if s.GoType == "[]byte" {
				template.IsRaw = true
			}
			return template, nil
		} else {
			ty, err := g.resolveSchema(content.Value.Schema, bodyStructName, true)
			if err != nil {
				panic(err)
			}
			if ty.GoType != "[]byte" {
				panic(fmt.Errorf("'%s'.%s: raw body must be []byte, got %s", "PATHTODO", "METHODTODO", ty.GoType))
			}
			return &RequestBodyTemplate{
				GoType: ty.GoType,
				IsRaw:  true,
			}, nil
		}
	}
	return nil, nil
}

// Handles generating a single response
// Note: x-type: empty is always extracted prior to calling this method, so it does not need to be considered.
func (g *Generator) genSingleResponse(baseName string, code int, response *oapi.ResponseOrRef) (*ResponseCaseTemplate, error) {
	if response.Ref != "" {
		path := strings.Replace(response.Ref, "#/components/responses/", "", 1)
		ref, ok := g.spec.Components.Responses.Get(path)
		if !ok {
			return nil, fmt.Errorf("failed to find response ref: %s", response.Ref)
		}

		return g.genSingleResponse(path, code, ref)
	}

	var resCase ResponseCaseTemplate
	resCase.Name = fmt.Sprintf("code%d", code)
	if len(response.Content) == 1 {
		single := response.Content[0]
		schema, err := g.resolveSchema(single.Value.Schema, baseName, false)
		if err != nil {
			panic(err)
		}

		resCase.GoType = schema.GoType
		resCase.Single = &SingleResponseTemplate{
			Code:        code,
			ContentType: single.Name,
			Name:        schema.GoType,
		}
	} else {
		// Generate a new model for this response type
		multiModelName := fmt.Sprintf("%s", baseName)
		fields := make([]*FieldInfo, len(response.Content))
		singles := make([]*SingleResponseTemplate, len(response.Content))

		for i, content := range response.Content {
			fieldName := contentTypeToFieldName(content.Name)
			schema, err := g.resolveSchema(content.Value.Schema, fieldName, false)
			if err != nil {
				panic(err)
			}

			singles[i] = &SingleResponseTemplate{
				Code:        code,
				ContentType: content.Name,
				Name:        schema.GoType,
			}
			fields[i] = &FieldInfo{
				Name: contentTypeToFieldName(content.Name),
				Type: schema.GoType,
			}
		}

		resCase.GoType = "*" + multiModelName
		resCase.Multi = &singles
		g.schemas2 = g.schemas2.With(fmt.Sprintf("#/components/schemas/%s", multiModelName), &TypeInfo{
			Name:   multiModelName,
			GoType: "*" + multiModelName,
			Struct: &StructType{
				Fields: fields,
			},
		})
	}

	return &resCase, nil
}

func contentTypeToFieldName(contentType string) string {
	sp := strings.Split(contentType, "/")
	contentType = sp[len(sp)-1]
	sp = strings.Split(contentType, ".")
	contentType = sp[len(sp)-1]
	contentType = strings.ReplaceAll(contentType, "_", "")
	contentType = strings.ReplaceAll(contentType, "-", "")
	return util.CamelToPascalCase(contentType)
}
