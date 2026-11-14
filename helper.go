package voyager

import (
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

func schemaForType(
	t reflect.Type,
	schemas openapi3.Schemas,
) (*openapi3.SchemaRef, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	value := reflect.New(t).Interface()

	return openapi3gen.NewSchemaRefForValue(
		value,
		schemas,
		openapi3gen.UseAllExportedFields(),
	)
}

func SchemaForType(
	t reflect.Type,
	schemas openapi3.Schemas,
) (*openapi3.SchemaRef, error) {
	return schemaForType(t, schemas)
}

func normalizeOpenAPIPath(path string) string {
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" +
				strings.TrimPrefix(part, ":") +
				"}"
		}
	}

	return strings.Join(parts, "/")
}
