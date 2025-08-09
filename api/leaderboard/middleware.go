package leaderboard

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var LastModified atomic.Value

func LastModifiedMiddleware(ctx *gin.Context) {
	clientMod := ctx.GetHeader("If-Modified-Since")
	lastMod, ok := LastModified.Load().(time.Time)

	if ok && clientMod != "" {
		clientTime, err := http.ParseTime(clientMod)
		if err == nil && !lastMod.After(clientTime) {
			ctx.Status(http.StatusNotModified)
			ctx.Abort()
			return
		}
	}
	if ok {
		ctx.Header("Last-Modified", lastMod.UTC().Format(http.TimeFormat))
	}
	ctx.Next()
}
