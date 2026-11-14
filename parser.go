package voyager

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

func inspectInput[I any]() ([]Parameter, []BodyField) {
	t := reflect.TypeFor[I]()

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	parameters := make([]Parameter, 0)
	body := make([]BodyField, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		parameterPtr := getParameters(field)
		if parameterPtr != nil {
			parameters = append(parameters, *parameterPtr)
		} else if name := field.Tag.Get("json"); name != "-" {
			body = append(body, BodyField{
				Name:        getJsonFieldName(name, field.Name),
				Type:        field.Type,
				Required:    isRequired(field),
				Example:     parseExample(field),
				Description: field.Tag.Get("description"),
			})
		}
	}

	return parameters, body
}

func getParameters(field reflect.StructField) *Parameter {
	example := parseExample(field)
	description := field.Tag.Get("description")

	switch {
	case field.Tag.Get("param") != "":
		return &Parameter{
			Name:        field.Tag.Get("param"),
			In:          ParameterPath,
			Required:    true,
			Type:        field.Type,
			Example:     example,
			Description: description,
		}

	case field.Tag.Get("query") != "":
		return &Parameter{
			Name:        field.Tag.Get("query"),
			In:          ParameterQuery,
			Required:    isRequired(field),
			Type:        field.Type,
			Example:     example,
			Description: description,
		}

	case field.Tag.Get("header") != "":
		return &Parameter{
			Name:        field.Tag.Get("header"),
			In:          ParameterHeader,
			Required:    isRequired(field),
			Type:        field.Type,
			Example:     example,
			Description: description,
		}
	}
	return nil
}

func getJsonFieldName(name string, fieldName string) string {
	if len(name) > 0 {
		return name
	}
	return fieldName
}

func isRequired(field reflect.StructField) bool {
	validate := field.Tag.Get("validate")

	for _, rule := range strings.Split(validate, ",") {
		if rule == "required" {
			return true
		}
	}

	return false
}

func parseExample(field reflect.StructField) any {
	raw, ok := field.Tag.Lookup("example")
	if !ok {
		return nil
	}

	t := field.Type

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return raw

	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err == nil {
			return value
		}

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:

		value, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			return value
		}

	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:

		value, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			return value
		}

	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			return value
		}

	case reflect.Slice,
		reflect.Array,
		reflect.Map,
		reflect.Struct:

		var value any

		if err := json.Unmarshal(
			[]byte(raw),
			&value,
		); err == nil {
			return value
		}
	}

	// A malformed example shouldn't prevent service startup.
	return raw
}
