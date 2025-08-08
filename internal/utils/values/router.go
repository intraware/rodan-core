package values

import (
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var router atomic.Value

func SetRouter(r *gin.Engine) {
	router.Store(r)
}

func GetRouter() *gin.Engine {
	val := router.Load()
	if val == nil {
		return nil
	}
	return val.(*gin.Engine)
}
