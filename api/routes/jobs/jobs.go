package jobs

import (
	"net/http"

	"github.com/diogovalentte/dashboard/api/job"
	"github.com/gin-gonic/gin"
)

func JobsRoutes(group *gin.RouterGroup) {
	{
		group.GET("/get_all", getAllJobs)
	}
}

func getAllJobs(c *gin.Context) {
	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get jobs"})
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobsList.GetJobs()})
}
