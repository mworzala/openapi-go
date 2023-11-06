package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/mworzala/openapi-go/internal/pkg/gen"
	"github.com/mworzala/openapi-go/pkg/oapi"
	"gopkg.in/yaml.v3"
)

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
}
