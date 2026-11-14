package voyager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type styleInput struct {
	ID   string `param:"id" json:"-" validate:"required"`
	Name string `query:"name" json:"-" validate:"required"`
}

type styleOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Trace string `json:"trace"`
}

func styleFunction(ctx *echo.Context, input styleInput) (styleOutput, error) {
	return styleOutput{ID: input.ID, Name: input.Name, Trace: ctx.Get("style-trace").(string)}, nil
}

type styleHandler struct{ info OperationInfo }

func (h *styleHandler) Handle(ctx *echo.Context, input styleInput) (styleOutput, error) {
	return styleFunction(ctx, input)
}

func (h *styleHandler) OpenAPI() OperationInfo { return h.info }

func TestHandlerStyles(t *testing.T) {
	functionRoutes := map[string]func(Router, string, func(*echo.Context, styleInput) (styleOutput, error), ...RouteOption) echo.RouteInfo{
		"GET": GET[styleInput, styleOutput], "POST": POST[styleInput, styleOutput],
		"PUT": PUT[styleInput, styleOutput], "PATCH": PATCH[styleInput, styleOutput],
		"DELETE": DELETE[styleInput, styleOutput],
	}
	structRoutes := map[string]func(Router, string, Handler[styleInput, styleOutput], ...RouteOption) echo.RouteInfo{
		"GET": GETHandler[styleInput, styleOutput], "POST": POSTHandler[styleInput, styleOutput],
		"PUT": PUTHandler[styleInput, styleOutput], "PATCH": PATCHHandler[styleInput, styleOutput],
		"DELETE": DELETEHandler[styleInput, styleOutput],
	}
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		for _, style := range []string{"function", "struct", "method"} {
			for _, required := range []bool{false, true} {
				name := method + "/" + style + "/public"
				if required {
					name = method + "/" + style + "/private"
				}
				t.Run(name, func(t *testing.T) {
					app := New("Handler styles", "1.0.0")
					info := OperationInfo{Summary: "Read user", Description: "Read by ID", Tags: []string{"Users"}, RequiredAuth: required}
					handler := &styleHandler{info: info}
					options := []RouteOption{WithMiddleware(func(next echo.HandlerFunc) echo.HandlerFunc {
						return func(ctx *echo.Context) error {
							ctx.Set("style-trace", "from middleware")
							return next(ctx)
						}
					})}
					path := "/handler-styles/" + name + "/:id"
					switch style {
					case "function":
						functionRoutes[method](app, path, styleFunction, append(options, WithOpenAPI(info))...)
					case "struct":
						structRoutes[method](app, path, handler, options...)
					case "method":
						functionRoutes[method](app, path, handler.Handle, append(options, WithOpenAPI(handler.OpenAPI()))...)
					}
					op := OpenAPI().Paths.Find(normalizeOpenAPIPath(path)).GetOperation(method)
					if op.Summary != info.Summary || op.Description != info.Description || !reflect.DeepEqual(op.Tags, info.Tags) {
						t.Fatalf("metadata lost: %+v", op)
					}
					if required {
						if op.Security == nil || len(*op.Security) != 1 {
							t.Fatalf("missing bearer requirement: %v", op.Security)
						}
						scopes, ok := (*op.Security)[0]["bearerAuth"]
						if !ok || scopes == nil || len(scopes) != 0 {
							t.Fatalf("invalid bearer requirement: %v", op.Security)
						}
					} else if op.Security != nil {
						t.Fatalf("public route requires authentication: %v", op.Security)
					}
					url := strings.Replace(path, ":id", "42", 1)
					rec := httptest.NewRecorder()
					app.ServeHTTP(rec, httptest.NewRequest(method, url+"?name=Alice", nil))
					if rec.Code != http.StatusOK {
						t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
					}
					var response Response[styleOutput]
					if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
						t.Fatal(err)
					}
					expected := styleOutput{ID: "42", Name: "Alice", Trace: "from middleware"}
					if response.Data != expected {
						t.Fatalf("response = %+v, want %+v", response.Data, expected)
					}
					rec = httptest.NewRecorder()
					app.ServeHTTP(rec, httptest.NewRequest(method, url, nil))
					if rec.Code != http.StatusBadRequest {
						t.Fatalf("missing required name: status %d", rec.Code)
					}
				})
			}
		}
	}
}

func TestMethodValueDoesNotInheritMetadata(t *testing.T) {
	app := echo.New()
	handler := &styleHandler{info: OperationInfo{Summary: "Private", RequiredAuth: true}}
	GET(app, "/method-without-metadata/:id", handler.Handle)
	op := OpenAPI().Paths.Find("/method-without-metadata/{id}").Get
	if op.Summary != "" || op.Security != nil {
		t.Fatalf("method value unexpectedly inherited struct metadata: %+v", op)
	}
}
