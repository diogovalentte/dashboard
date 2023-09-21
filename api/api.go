package api

import (
	"github.com/diogovalentte/dashboard/api/routes/notion"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")

	// Notion routes
	notion_group := v1.Group("/notion")
	{
		notion.GamesTrackerRoutes(notion_group)
	}

	return router
}
