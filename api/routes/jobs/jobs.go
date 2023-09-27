package jobs

import (
	"net/http"

	"github.com/diogovalentte/dashboard/api/job"
	"github.com/gin-gonic/gin"
)

func JobsRoutes(group *gin.RouterGroup) {
	{
		group.GET("/get_all", getAllJobs)
		group.DELETE("/delete_all", deleteAllJobs)
	}
}

func getAllJobs(c *gin.Context) {
	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't get jobs"})
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobsList.GetJobs()})
}

func deleteAllJobs(c *gin.Context) {
	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't get jobs list to delete"})
	}
	jobsList.DeleteAllJobs()
	c.JSON(http.StatusOK, gin.H{"message": "Jobs deleted with success"})
}
