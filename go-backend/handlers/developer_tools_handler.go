package handlers

import (
	"net/http"

	"pulse-control-plane/services"

	"github.com/gin-gonic/gin"
)

// DeveloperToolsHandler handles developer tools endpoints
type DeveloperToolsHandler struct {
	service *services.DeveloperToolsService
}

// NewDeveloperToolsHandler creates a new developer tools handler
func NewDeveloperToolsHandler() *DeveloperToolsHandler {
	return &DeveloperToolsHandler{
		service: services.NewDeveloperToolsService(),
	}
}

// GetPostmanCollection generates and returns a Postman collection
// @Summary Get Postman Collection
// @Description Download Postman collection for Pulse API
// @Tags Developer Tools
// @Produce json
// @Success 200 {object} services.PostmanCollection
// @Router /api/v1/developer/postman-collection [get]
func (h *DeveloperToolsHandler) GetPostmanCollection(c *gin.Context) {
	collection, err := h.service.GeneratePostmanCollection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename=pulse-api-collection.json")
	c.JSON(http.StatusOK, collection)
}

// DownloadGoSDK generates and downloads Go SDK
// @Summary Download Go SDK
// @Description Download auto-generated Go SDK
// @Tags Developer Tools
// @Produce application/zip
// @Success 200 {file} binary
// @Router /api/v1/developer/sdk/go [get]
func (h *DeveloperToolsHandler) DownloadGoSDK(c *gin.Context) {
	sdkZip, err := h.service.GenerateSDK("go")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename=pulse-go-sdk.zip")
	c.Data(http.StatusOK, "application/zip", sdkZip)
}

// DownloadJavaScriptSDK generates and downloads JavaScript SDK
// @Summary Download JavaScript SDK
// @Description Download auto-generated JavaScript SDK
// @Tags Developer Tools
// @Produce application/zip
// @Success 200 {file} binary
// @Router /api/v1/developer/sdk/javascript [get]
func (h *DeveloperToolsHandler) DownloadJavaScriptSDK(c *gin.Context) {
	sdkZip, err := h.service.GenerateSDK("javascript")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename=pulse-js-sdk.zip")
	c.Data(http.StatusOK, "application/zip", sdkZip)
}

// DownloadPythonSDK generates and downloads Python SDK
// @Summary Download Python SDK
// @Description Download auto-generated Python SDK
// @Tags Developer Tools
// @Produce application/zip
// @Success 200 {file} binary
// @Router /api/v1/developer/sdk/python [get]
func (h *DeveloperToolsHandler) DownloadPythonSDK(c *gin.Context) {
	sdkZip, err := h.service.GenerateSDK("python")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Header("Content-Disposition", "attachment; filename=pulse-python-sdk.zip")
	c.Data(http.StatusOK, "application/zip", sdkZip)
}

// GetOpenAPISpec returns OpenAPI 3.0 specification
// @Summary Get OpenAPI Specification
// @Description Get OpenAPI 3.0 spec for Pulse API
// @Tags Developer Tools
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/developer/openapi-spec [get]
func (h *DeveloperToolsHandler) GetOpenAPISpec(c *gin.Context) {
	spec, err := h.service.GenerateOpenAPISpec()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, spec)
}

// GetAPIDocumentation serves interactive API documentation
// @Summary API Documentation
// @Description Interactive API documentation (Swagger UI)
// @Tags Developer Tools
// @Produce html
// @Success 200 {string} html
// @Router /api/docs [get]
func (h *DeveloperToolsHandler) GetAPIDocumentation(c *gin.Context) {
	// Swagger UI HTML
	swaggerHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Pulse API Documentation</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
	<style>
		html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin:0; padding:0; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
	<script>
	window.onload = function() {
		const ui = SwaggerUIBundle({
			url: "/api/v1/developer/openapi-spec",
			dom_id: '#swagger-ui',
			deepLinking: true,
			presets: [
				SwaggerUIBundle.presets.apis,
				SwaggerUIStandalonePreset
			],
			plugins: [
				SwaggerUIBundle.plugins.DownloadUrl
			],
			layout: "StandaloneLayout"
		});
		window.ui = ui;
	};
	</script>
</body>
</html>
	`
	
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerHTML)
}
