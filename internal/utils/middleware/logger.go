package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		status := ctx.Writer.Status()
		fields := logrus.Fields{
			"type":     "http",
			"status":   status,
			"method":   ctx.Request.Method,
			"path":     ctx.Request.URL.Path,
			"errors":   ctx.Errors.String(),
			"time(µs)": time.Since(startTime).Microseconds(),
		}
		if msg, ok := ctx.Value("message").(string); ok && msg != "" {
			fields["message_context"] = msg
		}
		log := utils.Logger.WithFields(fields)
		switch status {
		case http.StatusNotFound:
			log.Trace("Path not found")
		case http.StatusInternalServerError:
			log.Warn("Internal server error")
		default:
			log.Info("Request processed")
		}
	}
}
