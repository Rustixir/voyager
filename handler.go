package voyager

import "github.com/labstack/echo/v5"

// Handler handles a typed request using Echo's context.
type Handler[I, O any] interface {
	Handle(*echo.Context, I) (O, error)
}

// HandlerFunc adapts a typed function to Handler.
type HandlerFunc[I, O any] func(*echo.Context, I) (O, error)

func (f HandlerFunc[I, O]) Handle(ctx *echo.Context, input I) (O, error) {
	return f(ctx, input)
}
