package voyager

type OpenAPIMetadata interface {
	OpenAPI() OperationInfo
}

type OperationInfo struct {
	Summary     string
	Description string
	Tags        []string
	// RequiredAuth documents HTTP bearer authentication; middleware must enforce it.
	RequiredAuth bool
}

type Describer interface {
	Description() string
}

type Summarizer interface {
	Summary() string
}

type Tagger interface {
	Tags() []string
}

func resolveOperationInfo(
	handler any,
) OperationInfo {
	if metadata, ok := handler.(OpenAPIMetadata); ok {
		return metadata.OpenAPI()
	}

	var info OperationInfo

	if summarizer, ok := handler.(Summarizer); ok {
		info.Summary = summarizer.Summary()
	}

	if describer, ok := handler.(Describer); ok {
		info.Description = describer.Description()
	}

	if tagger, ok := handler.(Tagger); ok {
		info.Tags = tagger.Tags()
	}

	return info
}

func ResolveOperationInfo(handler any) OperationInfo {
	return resolveOperationInfo(handler)
}
