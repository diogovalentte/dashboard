package api

import (
	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/routes/health_check"
	"github.com/diogovalentte/dashboard/api/routes/jobs"
	"github.com/diogovalentte/dashboard/api/routes/trackers"
	"github.com/gin-gonic/gin"
)

var jobsList *job.Jobs

func setRouterJobsList(jobsList *job.Jobs) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("JobsList", jobsList)
		c.Next()
	}
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	jobsList = job.NewJobsList()
	router.Use(setRouterJobsList(jobsList))

	v1 := router.Group("/v1")
	// Health check route
	{
		health_check.HealthCheckRoute(v1)
	}

	// Jobs routes
	jobsGroup := v1.Group("/jobs")
	{
		jobs.JobsRoutes(jobsGroup)
	}
	// Trackers routes
	trackersGroup := v1.Group("/trackers")
	{
		trackers.GamesTrackerRoutes(trackersGroup)
		trackers.MediasTrackerRoutes(trackersGroup)
	}

	return router
}
