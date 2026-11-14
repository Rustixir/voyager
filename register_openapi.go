package voyager

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func registerOpenAPI(operation Operation) error {
	openAPIMu.Lock()
	defer openAPIMu.Unlock()

	return addOperationToOpenAPI(
		operation,
	)
}

func addOperationToOpenAPI(
	operation Operation,
) error {
	op := openapi3.NewOperation()

	op.Summary = operation.Info.Summary
	op.Description = operation.Info.Description
	op.Tags = operation.Info.Tags

	if operation.Info.RequiredAuth {
		if doc.Components.SecuritySchemes == nil {
			doc.Components.SecuritySchemes = openapi3.SecuritySchemes{}
		}
		if _, exists := doc.Components.SecuritySchemes["bearerAuth"]; !exists {
			doc.Components.SecuritySchemes["bearerAuth"] = &openapi3.SecuritySchemeRef{
				Value: &openapi3.SecurityScheme{Type: "http", Scheme: "bearer"},
			}
		}
		op.Security = &openapi3.SecurityRequirements{
			{"bearerAuth": []string{}},
		}
	}

	if err := addParam(op, operation); err != nil {
		return err
	}

	if err := addRequestBody(
		operation,
		op,
	); err != nil {
		return err
	}

	if err := addResponse(
		operation,
		op,
	); err != nil {
		return err
	}

	doc.AddOperation(
		normalizeOpenAPIPath(operation.Path),
		operation.Method,
		op,
	)

	return nil
}

func addRequestBody(
	operation Operation,
	op *openapi3.Operation,
) error {
	if !methodAllowsBody(operation.Method) {
		return nil
	}

	if len(operation.Body) == 0 {
		return nil
	}

	schema := openapi3.NewObjectSchema()

	for _, field := range operation.Body {
		fieldSchema, err := schemaForType(
			field.Type,
			doc.Components.Schemas,
		)

		if err != nil {
			return fmt.Errorf(
				"voyager: build body field %q: %w",
				field.Name,
				err,
			)
		}

		if fieldSchema.Value != nil {
			fieldSchema.Value.Description = field.Description
			fieldSchema.Value.Example = field.Example
		}

		schema.Properties[field.Name] = fieldSchema

		if field.Required {
			schema.Required = append(
				schema.Required,
				field.Name,
			)
		}
	}

	op.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().
			WithRequired(true).
			WithJSONSchema(schema),
	}

	return nil
}

func addParam(
	op *openapi3.Operation,
	operation Operation,
) error {
	for _, parameter := range operation.Parameters {
		param, err := parameterToOpenAPI(
			parameter,
			doc.Components.Schemas,
		)

		if err != nil {
			return fmt.Errorf(
				"voyager: build parameter %q: %w",
				parameter.Name,
				err,
			)
		}

		op.AddParameter(param)
	}
	return nil
}

func addResponse(
	operation Operation,
	op *openapi3.Operation,
) error {
	if operation.Output == nil {
		return nil
	}

	outputSchema, err := schemaForType(
		operation.Output,
		doc.Components.Schemas,
	)
	if err != nil {
		return fmt.Errorf(
			"voyager: build response schema: %w",
			err,
		)
	}

	response := openapi3.NewResponse().
		WithDescription("Successful response").
		WithJSONSchemaRef(outputSchema)

	op.Responses = openapi3.NewResponses(
		openapi3.WithStatus(
			200,
			&openapi3.ResponseRef{
				Value: response,
			},
		),
	)
	return nil
}
