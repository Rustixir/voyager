package voyager

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestFunctionRouteOptions(t *testing.T) {
	app := New("Route options", "1.0.0")
	var order []int
	middleware := func(value int) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx *echo.Context) error {
				order = append(order, value)
				return next(ctx)
			}
		}
	}
	metadata := WithOpenAPI(OperationInfo{Summary: "Function route", Tags: []string{"Users"}, RequiredAuth: true})
	options := []RouteOption{metadata, WithMiddleware(middleware(1)), WithMiddleware(middleware(2))}
	GET(app, "/options/get/:id", pathTestHandler, options...)
	POST(app, "/options/post/:id", pathTestHandler, options...)
	PUT(app, "/options/put/:id", pathTestHandler, options...)
	PATCH(app, "/options/patch/:id", pathTestHandler, options...)
	DELETE(app, "/options/delete/:id", pathTestHandler, options...)
	for _, tc := range []struct{ method, path string }{
		{"GET", "get"}, {"POST", "post"}, {"PUT", "put"}, {"PATCH", "patch"}, {"DELETE", "delete"},
	} {
		op := OpenAPI().Paths.Find("/options/" + tc.path + "/{id}").GetOperation(tc.method)
		if op.Summary != "Function route" || !reflect.DeepEqual(op.Tags, []string{"Users"}) || op.Security == nil {
			t.Fatalf("metadata missing for %s: %+v", tc.method, op)
		}
		order = nil
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(tc.method, "/options/"+tc.path+"/42", nil))
		if rec.Code != 200 || !reflect.DeepEqual(order, []int{1, 2}) {
			t.Fatalf("%s: status %d, middleware order %v", tc.method, rec.Code, order)
		}
	}
}

func TestWithOpenAPIOverridesHandlerMetadata(t *testing.T) {
	app := echo.New()
	GETHandler(app, "/options/override", authDocHandler{required: true},
		WithOpenAPI(OperationInfo{Summary: "First", RequiredAuth: true}),
		WithOpenAPI(OperationInfo{Summary: "Last"}),
	)
	op := OpenAPI().Paths.Find("/options/override").Get
	if op.Summary != "Last" || op.Security != nil {
		t.Fatalf("last option did not replace handler metadata: %+v", op)
	}
}
