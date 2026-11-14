package voyager

import (
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type pathTestInput struct {
	ID string `param:"id"`
}

type pathTestOutput struct {
	OK bool `json:"ok"`
}

func pathTestHandler(
	ctx *echo.Context,
	input pathTestInput,
) (pathTestOutput, error) {
	return pathTestOutput{OK: true}, nil
}

// TestDocumentedPathIncludesGroupPrefix checks that an operation is documented
// under the path it is actually served on, prefixes included.
func TestDocumentedPathIncludesGroupPrefix(t *testing.T) {
	e := echo.New()

	handler := HandlerFunc[pathTestInput, pathTestOutput](pathTestHandler)

	api := e.Group("/api/v1")
	nested := api.Group("/admin")

	tests := []struct {
		name     string
		router   Router
		path     string
		expected string
	}{
		{
			name:     "no group",
			router:   e,
			path:     "/flat/:id",
			expected: "/flat/{id}",
		},
		{
			name:     "group prefix",
			router:   api,
			path:     "/products/:id",
			expected: "/api/v1/products/{id}",
		},
		{
			name:     "nested group prefix",
			router:   nested,
			path:     "/users/:id",
			expected: "/api/v1/admin/users/{id}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			routeInfo := GET(test.router, test.path, handler)

			// The route must still be served where Echo says it is.
			if routeInfo.Path != normalizeEchoPath(test.expected) {
				t.Errorf(
					"served path = %q, want %q",
					routeInfo.Path,
					normalizeEchoPath(test.expected),
				)
			}

			// ...and documented under that same path.
			if OpenAPI().Paths.Find(test.expected) == nil {
				t.Errorf(
					"path %q missing from OpenAPI document, have %v",
					test.expected,
					documentedPaths(),
				)
			}
		})
	}
}

// TestDocumentedPathFallsBackToRegisteredPath covers a Router implementation
// that reports no path back.
func TestDocumentedPathFallsBackToRegisteredPath(t *testing.T) {
	if got := documentedPath(echo.RouteInfo{}, "/fallback"); got != "/fallback" {
		t.Errorf("documentedPath = %q, want %q", got, "/fallback")
	}

	if got := documentedPath(
		echo.RouteInfo{Path: "/prefixed/thing"},
		"/thing",
	); got != "/prefixed/thing" {
		t.Errorf("documentedPath = %q, want %q", got, "/prefixed/thing")
	}
}

// normalizeEchoPath converts an OpenAPI-style path back to Echo's syntax.
func normalizeEchoPath(path string) string {
	replaced := ""

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '{':
			replaced += ":"
		case '}':
		default:
			replaced += string(path[i])
		}
	}

	return replaced
}

func documentedPaths() []string {
	paths := make([]string, 0)

	for path := range OpenAPI().Paths.Map() {
		paths = append(paths, path)
	}

	return paths
}

type bodyDocInput struct {
	ID   string `param:"id"`
	Note string `json:"note"`
}

type bodyDocOutput struct {
	Note string `json:"note"`
}

func bodyDocHandler(
	ctx *echo.Context,
	input bodyDocInput,
) (bodyDocOutput, error) {
	return bodyDocOutput{Note: input.Note}, nil
}

// TestRequestBodyDocumentedOnlyWhereBound keeps the generator and the binder in
// step: a body is documented for exactly the methods that read one.
func TestRequestBodyDocumentedOnlyWhereBound(t *testing.T) {
	e := echo.New()

	handler := HandlerFunc[bodyDocInput, bodyDocOutput](bodyDocHandler)

	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodGet, false},
		{http.MethodDelete, false},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			path := "/bodydoc/" + strings.ToLower(test.method) + "/:id"

			register(e, test.method, path, handler)

			item := OpenAPI().Paths.Find(
				"/bodydoc/" + strings.ToLower(test.method) + "/{id}",
			)
			if item == nil {
				t.Fatalf("path missing from document")
			}

			operation := item.GetOperation(test.method)
			if operation == nil {
				t.Fatalf("operation missing from document")
			}

			got := operation.RequestBody != nil
			if got != test.want {
				t.Errorf(
					"requestBody documented = %v, want %v (binds body = %v)",
					got,
					test.want,
					methodAllowsBody(test.method),
				)
			}
		})
	}
}
