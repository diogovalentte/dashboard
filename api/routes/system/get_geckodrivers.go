package system

import (
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SystemRoutes(group *gin.RouterGroup) {
	{
		group.GET("/get_geckodrivers", GetGeckoDriverInstances)
	}
}

func GetGeckoDriverInstances(c *gin.Context) {
	pool, err := scraping.NewGeckoDriverPool("", 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
		return
	}

	instancesAddr := pool.List()

	c.JSON(http.StatusOK, gin.H{"addresses": instancesAddr})
}
