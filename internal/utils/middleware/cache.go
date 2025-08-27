package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/utils/values"
)

func CacheMiddleware(ctx *gin.Context) {
	cache_time := values.GetConfig().App.CacheDuration
	ctx.Header("Cache-Control", fmt.Sprintf("public,max-age=%.0f", cache_time.Seconds()))
	ctx.Header("Expires", time.Now().Add(cache_time).Format(http.TimeFormat))
	ctx.Next()
}
