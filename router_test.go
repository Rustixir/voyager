package voyager

import (
	"testing"

	"github.com/labstack/echo/v5"
)

type documentedHandler struct{}

func (documentedHandler) Handle(ctx *echo.Context, input pathTestInput) (pathTestOutput, error) {
	return pathTestHandler(ctx, input)
}

func (documentedHandler) OpenAPI() OperationInfo {
	return OperationInfo{Summary: "Documented handler", Tags: []string{"example"}}
}

func TestRouteHelpers(t *testing.T) {
	app := echo.New()
	// Passing plain functions here checks generic type inference for every helper.
	GET(app, "/functions/get/:id", pathTestHandler)
	POST(app, "/functions/post/:id", pathTestHandler)
	PUT(app, "/functions/put/:id", pathTestHandler)
	PATCH(app, "/functions/patch/:id", pathTestHandler)
	DELETE(app, "/functions/delete/:id", pathTestHandler)

	handler := documentedHandler{}
	GETHandler(app, "/handlers/get/:id", handler)
	POSTHandler(app, "/handlers/post/:id", handler)
	PUTHandler(app, "/handlers/put/:id", handler)
	PATCHHandler(app, "/handlers/patch/:id", handler)
	DELETEHandler(app, "/handlers/delete/:id", handler)

	for _, method := range []struct{ name, path string }{
		{"GET", "get"}, {"POST", "post"}, {"PUT", "put"},
		{"PATCH", "patch"}, {"DELETE", "delete"},
	} {
		for _, prefix := range []string{"functions", "handlers"} {
			path := "/" + prefix + "/" + method.path + "/{id}"
			item := OpenAPI().Paths.Find(path)
			if item == nil {
				t.Fatalf("missing path %s", path)
			}
			operation := item.GetOperation(method.name)
			if operation == nil {
				t.Fatalf("missing operation %s %s", method.name, path)
			}
			if prefix == "handlers" && (operation.Summary != "Documented handler" || len(operation.Tags) != 1 || operation.Tags[0] != "example") {
				t.Fatalf("handler metadata lost for %s", method.name)
			}
		}
	}
}
