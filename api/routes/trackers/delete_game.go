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

func DeleteGame(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Delete game from Games Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	currentJob.SetStartingState("Processing game request")

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
		return
	}
	jobsList.AddJob(&currentJob)

	// Validate request
	var gameRequest DeleteGameRequest
	if err := c.ShouldBindJSON(&gameRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Delete game from DB
	currentJob.SetExecutingStateWithValue("Deleting game from DB", gameRequest.Name)
	err := deleteGame(&gameRequest)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	currentJob.SetCompletedState("Game deleted from DB")
	c.JSON(http.StatusOK, gin.H{"message": "Game deleted from DB"})
}

func deleteGame(gp *DeleteGameRequest) error {
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
    games_tracker
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

type DeleteGameRequest struct {
	Name string `json:"name" binding:"required"`
}
