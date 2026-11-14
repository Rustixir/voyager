package voyager

import (
	"log/slog"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
)

type ParameterLocation string

const (
	ParameterPath   ParameterLocation = "path"
	ParameterQuery  ParameterLocation = "query"
	ParameterHeader ParameterLocation = "header"
)

type Parameter struct {
	Name        string
	In          ParameterLocation
	Required    bool
	Type        reflect.Type
	Example     any
	Description string
}

type BodyField struct {
	Name        string
	Type        reflect.Type
	Required    bool
	Example     any
	Description string
}

type Operation struct {
	Method string
	Path   string

	Parameters []Parameter
	Body       []BodyField

	Output reflect.Type

	Info OperationInfo
}

func register[I, O any](
	router Router,
	method string,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	config := routeConfig{info: resolveOperationInfo(handler)}
	for _, option := range options {
		option(&config)
	}
	routeInfo := router.Add(
		method,
		path,
		adopt(handler),
		config.middleware...,
	)

	parameters, body := inspectInput[I]()

	operation := Operation{
		Method:     method,
		Path:       documentedPath(routeInfo, path),
		Parameters: parameters,
		Body:       body,
		Output:     reflect.TypeFor[Response[O]](),
		Info:       config.info,
	}

	if err := registerOpenAPI(operation); err != nil {
		panic(err)
	}

	return routeInfo
}

func documentedPath(
	routeInfo echo.RouteInfo,
	path string,
) string {
	if routeInfo.Path != "" {
		return routeInfo.Path
	}

	slog.Info(
		"voyager: router reported no path; documenting the path as registered",
		"path", path,
	)

	return path
}

func parameterToOpenAPI(
	parameter Parameter,
	schemas openapi3.Schemas,
) (*openapi3.Parameter, error) {
	schema, err := schemaForType(
		parameter.Type,
		schemas,
	)
	if err != nil {
		return nil, err
	}

	if schema.Value != nil {
		if parameter.Example != nil {
			schema.Value.Example = parameter.Example
		}

		schema.Value.Description = parameter.Description
	}

	return &openapi3.Parameter{
		Name:        parameter.Name,
		In:          string(parameter.In),
		Description: parameter.Description,
		Required:    parameter.Required,
		Schema:      schema,
	}, nil
}
