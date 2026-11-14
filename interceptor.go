package voyager

import "github.com/labstack/echo/v5"

const interceptorContextKey = "__voyager_interceptors"

type Interceptor interface {
	Intercept(ctx *echo.Context, input any) error
}

type InterceptorFunc func(ctx *echo.Context, input any) error

func (f InterceptorFunc) Intercept(ctx *echo.Context, input any) error {
	return f(ctx, input)
}

// Intercept runs functions after binding and before validation.
func Intercept(functions ...func(*echo.Context, any) error) echo.MiddlewareFunc {
	interceptors := make([]Interceptor, len(functions))
	for i, function := range functions {
		interceptors[i] = InterceptorFunc(function)
	}
	return InterceptHandler(interceptors...)
}

// InterceptHandler registers struct interceptors.
func InterceptHandler(interceptors ...Interceptor) echo.MiddlewareFunc {
	if len(interceptors) == 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			current := interceptorsFromContext(ctx)

			if current == nil {
				current = make(
					[]Interceptor,
					0,
					len(interceptors),
				)
			}

			current = append(
				current,
				interceptors...,
			)

			ctx.Set(interceptorContextKey, current)
			return next(ctx)
		}
	}
}

func interceptorsFromContext(ctx *echo.Context) []Interceptor {
	interceptors, _ := ctx.Get(interceptorContextKey).([]Interceptor)
	return interceptors
}
