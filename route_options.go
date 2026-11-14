package voyager

import "github.com/labstack/echo/v5"

type routeConfig struct {
	info       OperationInfo
	middleware []echo.MiddlewareFunc
}

// RouteOption configures a typed route.
type RouteOption func(*routeConfig)

// WithOpenAPI replaces the route's documentation metadata, including metadata
// provided by a struct handler. When repeated, the last option wins.
func WithOpenAPI(info OperationInfo) RouteOption {
	tags := append([]string(nil), info.Tags...)
	return func(config *routeConfig) {
		config.info = info
		config.info.Tags = append([]string(nil), tags...)
	}
}

// WithMiddleware appends route middleware in the order provided.
func WithMiddleware(middleware ...echo.MiddlewareFunc) RouteOption {
	middleware = append([]echo.MiddlewareFunc(nil), middleware...)
	return func(config *routeConfig) {
		config.middleware = append(config.middleware, middleware...)
	}
}
