package trackers

import (
	"database/sql"
	"fmt"
	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"net/http"
	"path/filepath"
	"time"
)

func UpdateGame(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Update game in Games Tracker database",
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
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("IsValidDate", IsValidDate)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	var gameRequest UpdateGameRequest
	if err := c.ShouldBindJSON(&gameRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := SetStructDateFields(&gameRequest)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Update media on DB
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if !gameRequest.Wait {
		go updateGameTask(&currentJob, &gameRequest, configs, c, gameRequest.Wait)
		c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
	} else {
		updateGameTask(&currentJob, &gameRequest, configs, c, gameRequest.Wait)
	}
}

func updateGameTask(currentJob *job.Job, gameRequest *UpdateGameRequest, configs *util.Configs, c *gin.Context, wait bool) {
	currentJob.SetExecutingStateWithValue("Updating game on the DB", gameRequest.Name)
	err := updateGame(gameRequest, configs)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetCompletedState("Game updated on DB")
	if wait {
		c.JSON(http.StatusOK, gin.H{"message": "Game updated on DB"})
	}
}

func updateGame(gameRequest *UpdateGameRequest, configs *util.Configs) error {
	dbPath := filepath.Join(configs.Database.FolderPath, "trackers.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	stm, err := db.Prepare(`
UPDATE
	games_tracker
SET
	priority = ?,
	status = ?,
	stars = ?,
	purchased_or_gamepass = ?,
	started_date = ?,
	finished_dropped_date = ?,
	commentary = ?,
	release_date = ?
WHERE
   name = ?
`)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(
		gameRequest.Priority,
		gameRequest.Status,
		gameRequest.Stars,
		gameRequest.PurchasedGamePass,
		gameRequest.StartedDate,
		gameRequest.FinishedDroppedDate,
		gameRequest.Commentary,
		gameRequest.ReleaseDate,
		gameRequest.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

type UpdateGameRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Whether the requester wants to wait for the task to be done before responding
	Name                   string    `json:"name" binding:"required"`
	Priority               int       `json:"priority" binding:"required"`
	Status                 int       `json:"status" binding:"required"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	PurchasedGamePass      bool      `json:"purchased_or_gamepass" binding:"-"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	ReleaseDateStr         string    `json:"release_date" binding:"omitempty,IsValidDate"`
	ReleaseDate            time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (gr *UpdateGameRequest) GetStartedDateStr() string {
	return gr.StartedDateStr
}

func (gr *UpdateGameRequest) GetFinishedDroppedDateStr() string {
	return gr.FinishedDroppedDateStr
}

func (gr *UpdateGameRequest) GetReleaseDateStr() string {
	return gr.ReleaseDateStr
}

func (gr *UpdateGameRequest) SetStartedDate(startedDate time.Time) {
	gr.StartedDate = startedDate
}

func (gr *UpdateGameRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	gr.FinishedDroppedDate = finishedDroppedDate
}

func (gr *UpdateGameRequest) SetReleaseDate(releaseDate time.Time) {
	gr.ReleaseDate = releaseDate
}
