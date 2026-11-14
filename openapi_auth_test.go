package voyager

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
)

type authDocHandler struct{ required bool }

func (h authDocHandler) Handle(*echo.Context, struct{}) (struct{}, error) {
	return struct{}{}, nil
}

func (h authDocHandler) OpenAPI() OperationInfo {
	return OperationInfo{Summary: "Authentication example", RequiredAuth: h.required}
}

func TestRequiredAuthDocumentation(t *testing.T) {
	app := New("Authentication", "1.0.0")
	GETHandler(app, "/auth-doc/private", authDocHandler{required: true})
	GETHandler(app, "/auth-doc/public", authDocHandler{})
	data, err := OpenAPIJSON()
	if err != nil {
		t.Fatal(err)
	}
	document, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	scheme := document.Components.SecuritySchemes["bearerAuth"]
	if scheme == nil || scheme.Value == nil || scheme.Value.Type != "http" || scheme.Value.Scheme != "bearer" {
		t.Fatal("missing HTTP bearer security scheme")
	}
	var raw struct {
		Paths map[string]map[string]struct {
			Security []map[string][]string `json:"security"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	private := raw.Paths["/auth-doc/private"]["get"].Security
	if len(private) != 1 {
		t.Fatalf("unexpected security requirements: %v", private)
	}
	scopes, ok := private[0]["bearerAuth"]
	if !ok || scopes == nil || len(scopes) != 0 {
		t.Fatalf("invalid bearer requirement: %v", private)
	}
	if public := raw.Paths["/auth-doc/public"]["get"].Security; len(public) != 0 {
		t.Fatalf("public operation requires authentication: %v", public)
	}
	// Documentation alone must not enforce authentication at runtime.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/auth-doc/private", nil))
	if rec.Code != 200 {
		t.Fatalf("unexpected runtime authentication: %d", rec.Code)
	}
}
