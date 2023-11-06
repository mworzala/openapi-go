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
	QueryParams  []*ParamTemplate
	HeaderParams []*ParamTemplate
	Body         *RequestBodyTemplate

	// Response
	Response *ResponseTemplate
}

type RequestBodyTemplate struct {
	GoType string
}

type ParamTemplate struct {
	Name     string
	Required bool
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
			result.QueryParams = append(result.QueryParams, &ParamTemplate{
				Name:     param.Name,
				Required: param.Required,
			})
		case "header":
			result.HeaderParams = append(result.HeaderParams, &ParamTemplate{
				Name:     param.Name,
				Required: param.Required,
			})
		default:
			println("unsupported param type: " + param.In)
		}
	}

	if op.RequestBody != nil {
		for _, content := range op.RequestBody.Content {
			if content.Name != "application/json" {
				panic("only json body supported.")
			}

			bodyStructName := fmt.Sprintf("%sRequest", util.SnakeToPascalCase(result.Name))
			s, err := g.genSingleSchema(bodyStructName, content.Value.Schema)
			if err != nil {
				panic(err)
			}
			g.schemas[bodyStructName] = s
			result.Body = &RequestBodyTemplate{}
			result.Body.GoType = s.GoType
		}
	}

	if len(op.Responses) > 0 {
		var res ResponseTemplate
		for _, response := range op.Responses {
			code, err := strconv.Atoi(response.Name)
			if err != nil {
				return nil, fmt.Errorf("'%s'.%s: failed to parse response code '%s': %w", path, method, response.Name, err)
			}

			switch response.Value.Type {
			case "empty":
				res.EmptyCode = &code
				continue
			case "success":
				// Continue below
			default:
				continue
			}

			var resCase ResponseCaseTemplate
			res.Cases = append(res.Cases, &resCase)
			resCase.Name = fmt.Sprintf("code%d", code)

			if len(response.Value.Content) == 1 {
				single := response.Value.Content[0]
				schema, err := g.resolveSchemaToType(single.Value.Schema)
				if err != nil {
					panic(err)
				}

				resCase.GoType = schema
				resCase.Single = &SingleResponseTemplate{
					Code:        code,
					ContentType: single.Name,
					Name:        schema,
				}
			} else {

				// Generate a new model for this response type
				multiModelName := fmt.Sprintf("%sResponse%d", util.SnakeToPascalCase(result.Name), code)
				fields := make([]*FieldTemplate, len(response.Value.Content))
				singles := make([]*SingleResponseTemplate, len(response.Value.Content))

				for i, content := range response.Value.Content {
					schema, err := g.resolveSchemaToType(content.Value.Schema)
					if err != nil {
						panic(err)
					}

					singles[i] = &SingleResponseTemplate{
						Code:        code,
						ContentType: content.Name,
						Name:        schema,
					}
					fields[i] = &FieldTemplate{
						Name: contentTypeToFieldName(content.Name),
						Type: schema,
					}
				}

				resCase.GoType = "*" + multiModelName
				resCase.Multi = &singles
				g.schemas[multiModelName] = &SchemaTemplate{
					Name: multiModelName,
					Struct: &StructTemplate{
						Name:   multiModelName,
						Fields: fields,
					},
				}
			}

		}
		result.Response = &res
	}

	return &result, nil
}

func contentTypeToFieldName(contentType string) string {
	sp := strings.Split(contentType, "/")
	contentType = sp[len(sp)-1]
	sp = strings.Split(contentType, ".")
	contentType = sp[len(sp)-1]
	return util.SnakeToPascalCase(contentType)
}
