package health_check

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheckRoute(group *gin.RouterGroup) {
	group.GET("/health", healthCheck)
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
