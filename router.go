package voyager

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Router is implemented by both *echo.Echo and *echo.Group.
type Router interface {
	Add(
		method string,
		path string,
		handler echo.HandlerFunc,
		middleware ...echo.MiddlewareFunc,
	) echo.RouteInfo
}

// ========================================================================
// HTTP Operations
// ========================================================================

// GET registers a typed handler function.
func GET[I, O any](
	router Router,
	path string,
	handler func(*echo.Context, I) (O, error),
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodGet,
		path,
		HandlerFunc[I, O](handler),
		options...,
	)
}

// GETHandler registers a struct handler and preserves its OpenAPI metadata.
func GETHandler[I, O any](
	router Router,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodGet,
		path,
		handler,
		options...,
	)
}

// POST registers a typed handler function.
func POST[I, O any](
	router Router,
	path string,
	handler func(*echo.Context, I) (O, error),
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPost,
		path,
		HandlerFunc[I, O](handler),
		options...,
	)
}

// POSTHandler registers a struct handler and preserves its OpenAPI metadata.
func POSTHandler[I, O any](
	router Router,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPost,
		path,
		handler,
		options...,
	)
}

// PUT registers a typed handler function.
func PUT[I, O any](
	router Router,
	path string,
	handler func(*echo.Context, I) (O, error),
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPut,
		path,
		HandlerFunc[I, O](handler),
		options...,
	)
}

// PUTHandler registers a struct handler and preserves its OpenAPI metadata.
func PUTHandler[I, O any](
	router Router,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPut,
		path,
		handler,
		options...,
	)
}

// PATCH registers a typed handler function.
func PATCH[I, O any](
	router Router,
	path string,
	handler func(*echo.Context, I) (O, error),
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPatch,
		path,
		HandlerFunc[I, O](handler),
		options...,
	)
}

// PATCHHandler registers a struct handler and preserves its OpenAPI metadata.
func PATCHHandler[I, O any](
	router Router,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodPatch,
		path,
		handler,
		options...,
	)
}

// DELETE registers a typed handler function.
func DELETE[I, O any](
	router Router,
	path string,
	handler func(*echo.Context, I) (O, error),
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodDelete,
		path,
		HandlerFunc[I, O](handler),
		options...,
	)
}

// DELETEHandler registers a struct handler and preserves its OpenAPI metadata.
func DELETEHandler[I, O any](
	router Router,
	path string,
	handler Handler[I, O],
	options ...RouteOption,
) echo.RouteInfo {
	return register(
		router,
		http.MethodDelete,
		path,
		handler,
		options...,
	)
}
