package notion

import (
	"github.com/diogovalentte/dashboard/api/routes/notion/games_tracker"
	"github.com/gin-gonic/gin"
)

func GamesTrackerRoutes(group *gin.RouterGroup) {
	games_tracker_group := group.Group("/games_tracker")
	{
		games_tracker_group.POST("/add_game", games_tracker.AddGame)
	}
}
