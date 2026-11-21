package handler

import (
	"net/http"

	layoutengine "github.com/abeta-dev/layout-engine"
	"github.com/gin-gonic/gin"
)

const scalarDocsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Abeta Layout Engine — API Docs</title>
  <style>body { margin: 0; }</style>
</head>
<body>
  <script id="api-reference" data-url="/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

func registerOpenAPIRoutes(engine *gin.Engine, serviceName string) {
	spec := map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Abeta Dynamic Layout Engine API",
			"version":     "1.0.0",
			"description": "Quick-commerce multi-tenant Server-Driven UI (SDUI) Layout Engine emitting v1 abeta-dev/kmp PageResponse contracts with rule targeting, time-window scheduling, dynamic slotting, immutable snapshots, rollback, fallback protection, and embedded Visual Layout Studio (/studio).",
		},
		"servers": []map[string]string{{"url": "/"}},
		"paths": map[string]any{
			"/studio":        map[string]any{"get": map[string]any{"summary": "Visual Layout Studio SPA", "tags": []string{"Studio UI"}}},
			"/llms.txt":      map[string]any{"get": map[string]any{"summary": "Plain-text llmstxt.org repository index", "tags": []string{"Docs & Discovery"}}},
			"/llms-full.txt": map[string]any{"get": map[string]any{"summary": "Plain-text full LLM engineering reference", "tags": []string{"Docs & Discovery"}}},
			"/api/v1/screens": map[string]any{
				"get": map[string]any{"summary": "List published dynamic screens (KMP-compatible)", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/screens/{screenId}": map[string]any{
				"get": map[string]any{"summary": "Resolve dynamic SDUI page by screenId (KMP drop-in)", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/pages/{slug}": map[string]any{
				"get": map[string]any{"summary": "Resolve dynamic SDUI page with optional ?include_trace=true", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/config/splash": map[string]any{
				"get": map[string]any{"summary": "Resolve rule-targeted splash configuration", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/catalog/tokens": map[string]any{
				"get": map[string]any{"summary": "Resolve design token catalog + runtime overrides", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/actions/execute": map[string]any{
				"post": map[string]any{"summary": "Execute SDUI action triggers (all 18 v1 action types)", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/orders/express": map[string]any{
				"post": map[string]any{"summary": "Place express quick-commerce order with celebration follow-up", "tags": []string{"Runtime SDUI"}},
			},
			"/api/v1/studio/tenants":                 map[string]any{"get": map[string]any{"summary": "List tenants", "tags": []string{"Studio CMS"}}, "post": map[string]any{"summary": "Create/update tenant", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/templates":               map[string]any{"get": map[string]any{"summary": "List 14 KMP snippet templates", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/widgets":                 map[string]any{"get": map[string]any{"summary": "List widgets", "tags": []string{"Studio CMS"}}, "post": map[string]any{"summary": "Create/update widget", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts":                 map[string]any{"get": map[string]any{"summary": "List layouts", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}":          map[string]any{"get": map[string]any{"summary": "Get layout draft", "tags": []string{"Studio CMS"}}, "put": map[string]any{"summary": "Create/update layout draft & slots", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/validate": map[string]any{"post": map[string]any{"summary": "Validate layout draft before publishing", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/publish":  map[string]any{"post": map[string]any{"summary": "Publish immutable versioned snapshot + changelog diff", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/versions": map[string]any{"get": map[string]any{"summary": "List historical published snapshots", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/rollback": map[string]any{"post": map[string]any{"summary": "Rollback layout to any historical snapshot version", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/fallback": map[string]any{"post": map[string]any{"summary": "Promote latest published version as guaranteed fallback", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/layouts/{slug}/preview":  map[string]any{"post": map[string]any{"summary": "Preview draft or published layout with simulated context & trace", "tags": []string{"Studio CMS"}}},
			"/api/v1/studio/rules":                   map[string]any{"get": map[string]any{"summary": "List targeting rules", "tags": []string{"Rules & Slotting"}}, "post": map[string]any{"summary": "Create/update targeting rule", "tags": []string{"Rules & Slotting"}}},
			"/api/v1/studio/slotting":                map[string]any{"get": map[string]any{"summary": "List dynamic slotting rules", "tags": []string{"Rules & Slotting"}}, "post": map[string]any{"summary": "Create/update slotting rule", "tags": []string{"Rules & Slotting"}}},
			"/api/v1/studio/splash":                  map[string]any{"get": map[string]any{"summary": "List splash configs", "tags": []string{"Config & Audit"}}, "post": map[string]any{"summary": "Create/update splash config", "tags": []string{"Config & Audit"}}},
			"/api/v1/studio/tokens":                  map[string]any{"get": map[string]any{"summary": "List token overrides", "tags": []string{"Config & Audit"}}, "post": map[string]any{"summary": "Create/update token override", "tags": []string{"Config & Audit"}}},
			"/api/v1/studio/changelog":               map[string]any{"get": map[string]any{"summary": "List immutable audit changelog with field diffs", "tags": []string{"Config & Audit"}}},
		},
		"x-service": serviceName,
	}

	engine.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, spec)
	})
	engine.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarDocsHTML))
	})
	engine.GET("/llms.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", layoutengine.LLMsTxt)
	})
	engine.GET("/llms-full.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", layoutengine.LLMsFullTxt)
	})
}
