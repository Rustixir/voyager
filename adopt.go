package voyager

import (
	"github.com/labstack/echo/v5"
	"net/http"
)

type Response[T any] struct {
	Data T `json:"data"`
}

func adopt[I, O any](handler Handler[I, O]) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		var input I
		if err := bindInput(ctx, &input); err != nil {
			return err
		}
		for _, interceptor := range interceptorsFromContext(ctx) {
			if err := interceptor.Intercept(ctx, &input); err != nil {
				return err
			}
		}
		if err := validateInput(ctx, &input); err != nil {
			return err
		}
		output, err := handler.Handle(ctx, input)
		if err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, Response[O]{Data: output})
	}
}
func bindInput(
	ctx *echo.Context,
	input any,
) error {
	if err := echo.BindPathValues(ctx, input); err != nil {
		return echo.NewHTTPError(400, "invalid path values")
	}

	if err := echo.BindQueryParams(ctx, input); err != nil {
		return echo.NewHTTPError(400, "invalid query parameters")
	}

	if methodAllowsBody(ctx.Request().Method) {
		if err := echo.BindBody(ctx, input); err != nil {
			return echo.NewHTTPError(400, "invalid request body")
		}
	}

	if err := echo.BindHeaders(ctx, input); err != nil {
		return echo.NewHTTPError(400, "invalid request headers")
	}

	return nil
}

func methodAllowsBody(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func validateInput(
	ctx *echo.Context,
	input any,
) error {
	var err error
	if ctx.Echo().Validator != nil {
		err = ctx.Validate(input)
	}
	if err != nil {
		return validationError(err)
	}

	vl, ok := input.(Validator)
	if ok {
		if err = vl.Validate(); err != nil {
			return validationError(err)
		}
	}

	return nil
}

func validationError(err error) error { return echo.NewHTTPError(http.StatusBadRequest, err.Error()) }
