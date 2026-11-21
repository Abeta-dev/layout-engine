package handler

import (
	"io/fs"
	"net/http"

	studioassets "github.com/abeta-dev/layout-engine/internal/handler/studio_assets"
	"github.com/gin-gonic/gin"
)

const studioCSP = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self' data:"

func registerStudioUIRoutes(engine *gin.Engine) {
	serveAsset := func(filename, contentType string) gin.HandlerFunc {
		return func(c *gin.Context) {
			data, err := fs.ReadFile(studioassets.Files, filename)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			c.Header("Content-Security-Policy", studioCSP)
			c.Header("Cache-Control", "no-cache")
			c.Data(http.StatusOK, contentType, data)
		}
	}

	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/studio")
	})
	engine.GET("/studio", serveAsset("index.html", "text/html; charset=utf-8"))
	engine.GET("/studio/", serveAsset("index.html", "text/html; charset=utf-8"))
	engine.GET("/studio/app.css", serveAsset("app.css", "text/css; charset=utf-8"))
	engine.GET("/studio/app.js", serveAsset("app.js", "application/javascript; charset=utf-8"))
}
