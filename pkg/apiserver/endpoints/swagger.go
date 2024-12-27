package endpoints

import "github.com/go-openapi/spec"

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Supported Resources API",
			Description: "Resource for managing API Resources",
			Version:     "v0.0.0",
		},
	}
}
