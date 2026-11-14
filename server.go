package voyager

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type requestValidator struct{ validator *validator.Validate }

func (v requestValidator) Validate(input any) error { return v.validator.Struct(input) }

// New creates an HTTP server with validation, panic recovery and generated API docs.
// Register routes before starting the server.
func New(title, version string) *echo.Echo {
	e := echo.New()
	e.Validator = requestValidator{validator.New()}
	e.Use(middleware.Recover())
	ConfigureOpenAPI(title, version)
	RegisterOpenAPI(e)
	RegisterOpenAPIDoc(e)
	return e
}
