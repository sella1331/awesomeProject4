package gzipMiddleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	contentType := g.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		if g.writer == nil {
			g.Header().Set("Content-Encoding", "gzip")
			g.Header().Del("Content-Length")
			g.writer = gzip.NewWriter(g.ResponseWriter)
		}
		return g.writer.Write(data)
	}
	return g.ResponseWriter.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer func() {
				_ = gz.Close()
			}()
			c.Request.Body = gz
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gw := &gzipWriter{ResponseWriter: c.Writer}
		c.Writer = gw

		c.Next()

		if gw.writer != nil {
			defer func() {
				_ = gw.writer.Close()
			}()
		}
	}
}
