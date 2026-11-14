package voyager

import (
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	openAPIMu sync.RWMutex

	doc = &openapi3.T{
		OpenAPI: "3.0.3",

		Info: &openapi3.Info{
			Title:   "Voyager API",
			Version: "1.0.0",
		},

		Paths: openapi3.NewPaths(),

		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
	}
)

func ConfigureOpenAPI(
	title string,
	version string,
) {
	openAPIMu.Lock()
	defer openAPIMu.Unlock()

	doc.Info.Title = title
	doc.Info.Version = version
}

func OpenAPI() *openapi3.T {
	openAPIMu.RLock()
	defer openAPIMu.RUnlock()

	return doc
}
