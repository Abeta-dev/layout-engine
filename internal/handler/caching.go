package handler

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultCacheControl = "public, max-age=15"

// ETagMiddleware computes a SHA-256 ETag over successful GET responses and handles 304 Not Modified.
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		orig := c.Writer
		bw := newBufferedResponseWriter(orig)
		c.Writer = bw
		completed := false
		defer func() {
			if !completed {
				c.Writer = orig
			}
		}()

		c.Next()
		completed = true
		c.Writer = orig

		status := bw.Status()
		body := bw.buf.Bytes()
		if status < 200 || status >= 300 || len(body) == 0 {
			if bw.wroteHeader || len(body) > 0 {
				orig.WriteHeader(status)
			}
			if len(body) > 0 && bodyAllowedForStatus(status) {
				_, _ = orig.Write(body)
			}
			return
		}

		sum := sha256.Sum256(body)
		etag := fmt.Sprintf(`"%x"`, sum)
		h := orig.Header()
		h.Set("ETag", etag)
		h.Set("Cache-Control", defaultCacheControl)
		h.Del("Pragma")

		if ifNoneMatchMatches(c.GetHeader("If-None-Match"), etag) {
			h.Del("Content-Length")
			h.Del("Content-Type")
			orig.WriteHeader(http.StatusNotModified)
			return
		}
		orig.WriteHeader(status)
		_, _ = orig.Write(body)
	}
}

func ifNoneMatchMatches(headerVal, currentETag string) bool {
	trimmed := strings.TrimSpace(headerVal)
	if trimmed == "" {
		return false
	}
	if trimmed == "*" {
		return true
	}
	normCurrent := strings.TrimPrefix(currentETag, "W/")
	for _, part := range strings.Split(trimmed, ",") {
		cand := strings.TrimSpace(part)
		if cand == "*" || (cand != "" && strings.TrimPrefix(cand, "W/") == normCurrent) {
			return true
		}
	}
	return false
}
