package gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/mworzala/openapi-go/internal/pkg/util"
	"github.com/mworzala/openapi-go/pkg/oapi"
)

var (
	//go:embed model.tmpl
	modelTemplateRaw string
	//go:embed server.tmpl
	serverTemplateRaw string
	templateFuncs     = template.FuncMap{
		"CamelToPascal":            util.CamelToPascalCase,
		"SnakeToPascal":            util.SnakeToPascalCase,
		"DashToCamel":              util.DashToCamelCase,
		"FieldNameFromContentType": contentTypeToFieldName,
		"NoPtr": func(t string) string {
			if len(t) > 0 && t[0] == '*' {
				return t[1:]
			} else {
				return t
			}
		},
	}
)

// Generator represents a generator for a chi server given
// a set of openapi specs.
type Generator struct {
	pwd            string
	modelTemplate  *template.Template
	serverTemplate *template.Template

	// Single spec processing state

	// The spec currently being processed
	spec       *oapi.Spec
	specName   string
	apiVersion string
	// Set of schemas to be emitted (by absolute path from file, eg #/components/schemas/MySchema)
	schemas    map[string]*SchemaTemplate
	schemas2   oapi.MapSlice[TypeInfo]
	operations []*OperationTemplate
}

func New() (*Generator, error) {
	pwd, err := os.Getwd()

	modelTemplate, err := template.New("model.tmpl").Funcs(templateFuncs).Parse(modelTemplateRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model template: %w", err)
	}
	serverTemplate, err := template.New("server.tmpl").Funcs(templateFuncs).Parse(serverTemplateRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server template: %w", err)
	}

	return &Generator{
		pwd:            pwd,
		modelTemplate:  modelTemplate,
		serverTemplate: serverTemplate,

		schemas:  make(map[string]*SchemaTemplate),
		schemas2: make(oapi.MapSlice[TypeInfo], 0),
	}, nil
}

func (g *Generator) GenSpecSingle(name string, spec *oapi.Spec) {
	defer g.flush()

	if strings.HasSuffix(name, "_v2") {
		g.specName = strings.TrimSuffix(name, "_v2")
		g.apiVersion = "v2"
	} else {
		g.specName = name
		g.apiVersion = "v1"
	}
	g.spec = spec

	// Generate server
	for _, specOp := range spec.Paths {
		if specOp.Value == nil {
			panic("missing value!!")
		}
		if specOp.Value.IsEmpty() {
			continue
		}

		if specOp.Value.Get != nil {
			op, err := g.genOperation(specOp.Name, "get", specOp.Value.Get)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

		if specOp.Value.Put != nil {
			op, err := g.genOperation(specOp.Name, "put", specOp.Value.Put)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

		if specOp.Value.Post != nil {
			op, err := g.genOperation(specOp.Name, "post", specOp.Value.Post)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

		if specOp.Value.Delete != nil {
			op, err := g.genOperation(specOp.Name, "delete", specOp.Value.Delete)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

		if specOp.Value.Patch != nil {
			op, err := g.genOperation(specOp.Name, "patch", specOp.Value.Patch)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

		if specOp.Value.Trace != nil {
			op, err := g.genOperation(specOp.Name, "trace", specOp.Value.Trace)
			if err != nil {
				panic(err)
			}
			g.operations = append(g.operations, op)
		}

	}

	if g.spec.Components != nil {
		// Append any models which were not referenced from the spec
		for _, schema := range g.spec.Components.Schemas {
			fullName := fmt.Sprintf("#/components/schemas/%s", schema.Name)
			if _, ok := g.schemas2.Get(fullName); ok {
				continue // Already generated, skip
			}

			ti, err := g.resolveSchema(schema.Value, schema.Name, false)
			if err != nil {
				panic(fmt.Errorf("failed to generate schema %s: %w", schema.Name, err))
			}
			g.schemas2 = g.schemas2.With(fullName, ti)
		}
	}
}

func (g *Generator) flush() {

	// Path name override
	basePath := "/" + g.specName
	if g.spec.Info.BasePath != nil {
		basePath = *g.spec.Info.BasePath
	}

	// Write the server to server file
	serverFile := path.Join(g.pwd, fmt.Sprintf("%s_server.gen.go", g.specName))
	serverContext := &ServerTemplate{
		Package: g.apiVersion, Name: g.specName,
		BasePath: fmt.Sprintf("/%s%s", g.apiVersion, basePath), Operations: g.operations, UseFx: true}
	if err := execTemplateToFile(g.serverTemplate, serverContext, serverFile); err != nil {
		panic(fmt.Errorf("failed to execute server template: %w", err))
	}

	// Write models to model file
	schemas := make([]*TypeInfo, 0, len(g.schemas2))
	for _, schema := range g.schemas2 {
		schemas = append(schemas, schema.Value)
	}

	modelFile := path.Join(g.pwd, fmt.Sprintf("%s_model.gen.go", g.specName))
	context := &ModelTemplate{Package: g.apiVersion, Schemas: schemas}
	if err := execTemplateToFile(g.modelTemplate, context, modelFile); err != nil {
		panic(fmt.Errorf("failed to execute model template: %w", err))
	}

	// Cleanup
	g.spec = nil
	g.specName = ""
	g.schemas = make(map[string]*SchemaTemplate)
	g.operations = nil

	// Some fun stats
	println("Generated", serverContext.BasePath)
	println("Total operations:", len(serverContext.Operations))
	println("Total schemas:", len(schemas))
}

func execTemplateToFile(t *template.Template, context interface{}, fileName string) error {
	out := new(bytes.Buffer)

	// Execute the template and format the code
	err := t.Execute(out, context)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	//println(string(out.Bytes()))
	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format source: %w", err)
	}

	// Write the code to disk (replacing existing)
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fileName, err)
	}
	defer f.Close()
	if _, err = f.Write(formatted); err != nil {
		return fmt.Errorf("failed to write file data %s: %w", fileName, err)
	}
	//if _, err = f.Write(out.Bytes()); err != nil {
	//	return fmt.Errorf("failed to write file data %s: %w", fileName, err)
	//}

	// Run goimports (todo i would like this to not break without it installed)
	cmd := exec.Command("goimports", "-w", fileName)
	if err = cmd.Run(); err != nil {
		log.Fatalf("goimports failed: %v", err)
	}

	return nil
}
