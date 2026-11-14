package voyager

import "github.com/labstack/echo/v5"

func OpenAPIJSON() ([]byte, error) {
	openAPIMu.RLock()
	defer openAPIMu.RUnlock()

	return doc.MarshalJSON()
}

func RegisterOpenAPI(e *echo.Echo) {
	e.GET("/openapi.json", func(ctx *echo.Context) error {
		data, err := OpenAPIJSON()
		if err != nil {
			return err
		}

		return ctx.Blob(
			200,
			"application/json",
			data,
		)
	})
}

func RegisterOpenAPIDoc(e *echo.Echo) {
	e.GET("/docs", func(ctx *echo.Context) error {
		html := `
			<!DOCTYPE html>
			<html>
			<head>
			  <meta charset="UTF-8">
			  <title>API Docs</title>
			
			  <link
				rel="stylesheet"
				href="https://unpkg.com/swagger-ui-dist/swagger-ui.css"
			  />
			</head>
			
			<body>
			  <div id="swagger-ui"></div>
			
			  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
			
			  <script>
				SwaggerUIBundle({
				  url: "openapi.json",
				  dom_id: "#swagger-ui",
				  deepLinking: true,
				  displayRequestDuration: true,
				  tryItOutEnabled: true
				})
			  </script>
			</body>
			</html>
		`
		return ctx.HTML(200, html)
	})
}
