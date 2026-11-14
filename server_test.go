package voyager

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/getkin/kin-openapi/openapi3"
)

type createInput struct {
	ID   string `param:"id" json:"-"`
	Name string `json:"name" validate:"required"`
}
type createOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func createHandler(ctx *echo.Context, in createInput) (createOutput, error) {
	if ctx.Request().Context() == nil {
		return createOutput{}, errors.New("missing request context")
	}
	return createOutput{ID: in.ID, Name: in.Name}, nil
}

func TestServerBindingValidationAndDocumentation(t *testing.T) {
	app := New("Test API", "1.2.3")
	POST(app, "/integration/:id", createHandler)
	for _, tc := range []struct {
		body   string
		status int
	}{
		{`{"name":"Voyager","id":"spoofed"}`, 200},
		{`{}`, 400},
		{`{"name":`, 400},
	} {
		req := httptest.NewRequest(http.MethodPost, "/integration/42", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("body %s: status %d: %s", tc.body, rec.Code, rec.Body.String())
		}
		if tc.status == 200 && !strings.Contains(rec.Body.String(), `"id":"42"`) {
			t.Fatal(rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/openapi.json", nil))
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	doc, err := openapi3.NewLoader().LoadFromData(rec.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if err = doc.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if doc.Info.Title != "Test API" || doc.Paths.Value("/integration/{id}").Post == nil {
		t.Fatal("missing generated operation")
	}
	var raw map[string]any
	if err = json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/docs", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "SwaggerUIBundle") {
		t.Fatal("missing Swagger UI")
	}
}

type customInput struct{}

func (*customInput) Validate() error { return context.Canceled }
func TestCustomValidationWithoutEchoValidator(t *testing.T) {
	app := New("Custom validation", "1.0.0")
	app.Validator = nil
	POST(app, "/custom-validation", func(*echo.Context, customInput) (struct{}, error) {
		t.Fatal("invalid input reached handler")
		return struct{}{}, nil
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("POST", "/custom-validation", nil))
	if rec.Code != 400 {
		t.Fatalf("status %d", rec.Code)
	}
}
