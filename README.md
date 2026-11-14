# Voyager

Typed HTTP handlers for Go, built on Echo v5.

Voyager removes two repetitive parts of building an API: writing request-binding code and maintaining Swagger documentation. Define your input, output, and handler once. Voyager binds and validates requests, runs your handler, and generates an OpenAPI 3 specification with Swagger UI.

## Requirements

Go 1.25 or newer.

The module path is currently `voyager`. The examples below use that local import path. A public installation command will be available once the module has a public repository path.

## Quick start

```go
package main

import (
    "log"

    "voyager"

    "github.com/labstack/echo/v5"
)

type GreetInput struct {
    Name     string `param:"name" json:"-" validate:"required" description:"Name to greet"`
    Language string `query:"language" json:"-" description:"Optional greeting language"`
}

type GreetOutput struct {
    Message string `json:"message"`
}

func greet(ctx *echo.Context, input GreetInput) (GreetOutput, error) {
    greeting := "Hello"
    if input.Language == "fa" {
        greeting = "سلام"
    }
    return GreetOutput{Message: greeting + " " + input.Name}, nil
}

func main() {
    app := voyager.New("Hello API", "1.0.0")
    voyager.GET(app, "/hello/:name", greet)
    log.Fatal(app.Start(":8080"))
}
```

Run the included example from the repository root:

```sh
go run ./examples/hello
```

Then send a request:

```sh
curl 'http://localhost:8080/hello/world?language=fa'
```

```json
{"data":{"message":"سلام world"}}
```

Open [Swagger UI](http://localhost:8080/docs) to explore the API, or retrieve the [OpenAPI specification](http://localhost:8080/openapi.json).

`voyager.New` returns an `*echo.Echo` with validation, panic recovery, and documentation routes configured.

## Request binding

Use exported fields in a struct input. Tags tell Voyager where to read each value:

| Tag | Source | Methods |
| --- | --- | --- |
| `param:"id"` | Path parameter, such as `/users/:id` | All supported methods |
| `query:"page"` | Query string | All supported methods |
| `header:"X-Request-ID"` | Request header | All supported methods |
| `json:"name"` | JSON body | POST, PUT, PATCH |
| `form:"name"` | Form body | POST, PUT, PATCH |

For example, an update request can combine path and body fields:

```go
type UpdateUserInput struct {
    ID   string `param:"id" json:"-" validate:"required"`
    Name string `json:"name" validate:"required" description:"Display name" example:"Alice"`
}
```

A request to `PUT /users/42` with `{"name":"Alice"}` binds `ID` to `"42"` and `Name` to `"Alice"`. Set `Content-Type: application/json` when sending JSON.

Binding runs in this order: path, query, body, headers. Later sources can overwrite earlier values. Use one source per field and add `json:"-"` to path, query, and header fields to prevent JSON bodies from overwriting them. JSON binding can also populate exported fields without a `json` tag, so keep request structs limited to fields clients may supply.

Use struct inputs rather than pointer inputs. For a route with no input, use `struct{}`.

## Validation and errors

`voyager.New` enables go-playground/validator tags:

```go
type CreateUserInput struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}
```

For custom validation, implement `Validate() error` on the input pointer:

```go
func (input *CreateUserInput) Validate() error {
    if input.Name == input.Email {
        return errors.New("name and email must be different")
    }
    return nil
}
```

Import `errors` for this example. Custom validation runs after tag validation succeeds. If validation fails, the handler is not called.

| Result | HTTP behavior |
| --- | --- |
| Handler returns an output and `nil` | HTTP 200 with `{"data":...}` |
| Binding or validation fails | HTTP 400 |
| Handler returns an error | Passed to Echo's error handler |

Return an Echo HTTP error when you need a specific error status:

```go
return GreetOutput{}, echo.NewHTTPError(http.StatusNotFound, "user not found")
```

This example uses `github.com/labstack/echo/v5` and `net/http`. Successful responses currently always use HTTP 200 and the `data` envelope.

## Routes, groups, and middleware

Register typed handlers with `voyager.GET`, `POST`, `PUT`, `PATCH`, or `DELETE`. Each accepts an Echo instance or group, a path, a typed function, and optional route options. Go infers the input and output types from the function.

```go
api := app.Group("/api/v1")
voyager.GET(api, "/hello/:name", greet)
```

The route is served at `/api/v1/hello/:name` and documented as `/api/v1/hello/{name}`. Nested group prefixes are included in the generated specification.

Echo middleware works normally through `app.Use` and groups. For route-specific middleware, use `WithMiddleware`:

```go
voyager.GET(app, "/hello/:name", greet,
    voyager.WithMiddleware(middleware.Logger()),
)
```

This example uses `github.com/labstack/echo/v5/middleware`. Multiple `WithMiddleware` options append middleware in the order provided. Direct middleware arguments on Voyager route helpers must now be wrapped in `WithMiddleware`.

### Function handlers

Pass a function directly to `GET`, `POST`, `PUT`, `PATCH`, or `DELETE`. Use `WithOpenAPI` to add documentation metadata:

```go
func greet(ctx *echo.Context, input GreetInput) (GreetOutput, error) {
    return GreetOutput{Message: "Hello " + input.Name}, nil
}

// Inside main, after creating app:
// voyager.GET(app, "/hello/:name", greet,
//     voyager.WithOpenAPI(voyager.OperationInfo{
//         Summary: "Greet someone",
//         Tags:    []string{"Greetings"},
//     }),
// )
```

### Struct handlers

Use a struct handler to group behavior and dependencies. Implement `Handle` and optionally `OpenAPI`:

```go
type GreetHandler struct{}

func (GreetHandler) Handle(ctx *echo.Context, input GreetInput) (GreetOutput, error) {
    return greet(ctx, input)
}

func (GreetHandler) OpenAPI() voyager.OperationInfo {
    return voyager.OperationInfo{
        Summary:     "Greet someone",
        Description: "Returns a greeting in the requested language.",
        Tags:        []string{"Greetings"},
    }
}

// Register inside main, after creating app:
// voyager.GETHandler(app, "/hello/:name", GreetHandler{})
```

A struct handler implements `Handle(*echo.Context, Input) (Output, error)`. Register it with `GETHandler`, `POSTHandler`, `PUTHandler`, `PATCHHandler`, or `DELETEHandler` to preserve its documentation metadata. The function helpers (`GET`, `POST`, and so on) accept ordinary functions directly.

| HTTP method | Function handler | Struct handler |
| --- | --- | --- |
| GET | `voyager.GET` | `voyager.GETHandler` |
| POST | `voyager.POST` | `voyager.POSTHandler` |
| PUT | `voyager.PUT` | `voyager.PUTHandler` |
| PATCH | `voyager.PATCH` | `voyager.PATCHHandler` |
| DELETE | `voyager.DELETE` | `voyager.DELETEHandler` |

Both styles accept `WithOpenAPI` and `WithMiddleware` options. You can also pass a struct method as a function:

```go
handler := GreetHandler{}
voyager.GET(app, "/hello/:name", handler.Handle)
```

Passing `handler.Handle` does not carry the struct's `OpenAPI()` metadata. Use `GETHandler` to read it automatically, or provide `WithOpenAPI(handler.OpenAPI())` explicitly.

### Request context

Handlers and interceptors receive `*echo.Context` directly. Use `ctx.Request().Context()` for cancellation and deadlines, and Echo's context methods for headers, cookies, and other HTTP operations.

### Interceptors

Interceptors run after binding and before validation. They receive a pointer to the input and can inspect or modify it, or return an error to stop the request.

```go
app.Use(voyager.Intercept(func(ctx *echo.Context, input any) error {
    if greeting, ok := input.(*GreetInput); ok && greeting.Language == "" {
        greeting.Language = "en"
    }
    return nil
}))
```

For struct interceptors that implement `Intercept(*echo.Context, any) error`, use `voyager.InterceptHandler(interceptor)`. You can pass multiple functions to `Intercept` or multiple struct interceptors to `InterceptHandler`; they run in registration order.

## OpenAPI documentation

Voyager generates OpenAPI 3.0.3 documentation when you register typed routes. No separate generator command or Swagger comment annotations are required.

The generated document includes route paths, HTTP methods, path/query/header parameters, JSON request bodies, and successful response schemas with the `data` envelope.

| Field tag | Documentation |
| --- | --- |
| `validate:"required"` | Marks a parameter or body field as required |
| `description:"..."` | Adds a parameter or body-field description |
| `example:"..."` | Adds a parameter or body-field example |

Path parameters are always documented as required. Examples are parsed according to the field type when possible. Runtime validation supports more rules than the generator reflects in the specification.

Implement `OpenAPI() voyager.OperationInfo` on a handler to add its summary, description, and tags. Alternatively, implement `Summary() string`, `Description() string`, or `Tags() []string`. When `OpenAPI()` is present, it takes precedence.

### Authentication

Set `RequiredAuth: true` in a struct handler's OpenAPI metadata to document HTTP bearer authentication:

```go
func (GetProfileHandler) OpenAPI() voyager.OperationInfo {
    return voyager.OperationInfo{
        Summary:      "Get profile",
        Tags:         []string{"Users"},
        RequiredAuth: true,
    }
}
```

Register the handler with `voyager.GETHandler(app, "/profile", GetProfileHandler{})`.

For function handlers, pass the same metadata with `WithOpenAPI`:

```go
voyager.GET(app, "/profile", getProfile,
    voyager.WithOpenAPI(voyager.OperationInfo{
        Summary:      "Get profile",
        Tags:         []string{"Users"},
        RequiredAuth: true,
    }),
)
```

`WithOpenAPI` works with all function and struct route helpers. It replaces the entire metadata value, including metadata returned by a struct handler. If provided more than once, the last option wins.

Voyager adds a `bearerAuth` security scheme and marks the operation as requiring it, allowing you to supply a bearer token in Swagger UI. The field defaults to `false`, so other operations do not receive this requirement. Authentication must be enforced by your middleware; the flag only controls documentation.

To export the generated specification programmatically:

```go
data, err := voyager.OpenAPIJSON()
```

## Development

Run checks from the repository root:

```sh
go test ./...
go test -race ./...
go vet ./...
```

Run the example with `go run ./examples/hello` or `make run`.
