package trackers

import (
	"database/sql"
	"fmt"
	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"time"
)

func DeleteMedia(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Delete media from Medias Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	currentJob.SetStartingState("Processing media request")

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
		return
	}
	jobsList.AddJob(&currentJob)

	// Validate request
	var mediaRequest DeleteMediaRequest
	if err := c.ShouldBindJSON(&mediaRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Delete media from DB
	currentJob.SetExecutingStateWithValue("Deleting media from DB", mediaRequest.Name)
	err := deleteMedia(&mediaRequest)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	currentJob.SetCompletedState("Media deleted from DB")
	c.JSON(http.StatusOK, gin.H{"message": "Media deleted from DB"})
}

func deleteMedia(gp *DeleteMediaRequest) error {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		return err
	}
	dbPath := filepath.Join(configs.Database.FolderPath, "trackers.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
DELETE FROM
    medias_tracker
WHERE
    name = ?;
  `)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(
		gp.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

type DeleteMediaRequest struct {
	Name string `json:"name" binding:"required"`
}
