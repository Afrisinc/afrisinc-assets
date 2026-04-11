package handler

import (
	_ "embed"
	"net/http"
)

//go:embed docs/openapi.yaml
var openAPISpec string

// DocsHandler serves the OpenAPI/Swagger specification.
type DocsHandler struct{}

// NewDocsHandler creates a new DocsHandler.
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// OpenAPI handles GET /api/docs/openapi.yaml
func (h *DocsHandler) OpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(openAPISpec)) //nolint:errcheck
}

// SwaggerHTML handles GET /api/docs or /api/docs/
func (h *DocsHandler) SwaggerHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML)) //nolint:errcheck
}

var swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Afrisinc Assets API - Swagger UI</title>
	<link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
	<style>
		html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin: 0; padding: 0; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js" charset="UTF-8"><\/script>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-standalone-preset.js" charset="UTF-8"><\/script>
	<script>
		const ui = SwaggerUIBundle({
			url: "/api/docs/openapi.yaml",
			dom_id: "#swagger-ui",
			presets: [
				SwaggerUIBundle.presets.apis,
				SwaggerUIStandalonePreset
			],
			layout: "StandaloneLayout",
			deepLinking: true,
			showCommonExtensions: true,
			showExtensions: true,
			docExpansion: "list",
			defaultModelsExpandDepth: 1,
			defaultModelExpandDepth: 1,
			requestInterceptor: (request) => {
				const apiKey = localStorage.getItem("apiKey");
				if (apiKey) {
					request.headers["X-API-Key"] = apiKey;
				}
				return request;
			}
		});
	<\/script>
</body>
</html>`
