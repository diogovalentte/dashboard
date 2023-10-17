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

func UpdateMedia(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Update media in Medias Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
		return
	}
	jobsList.AddJob(&currentJob)
	currentJob.SetStartingState("Processing media request")

	// Validate request
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("IsValidDate", IsValidDate)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	var mediaRequest UpdateMediaRequest
	if err := c.ShouldBindJSON(&mediaRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := SetStructDateFields(&mediaRequest)
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

	if !mediaRequest.Wait {
		go updateMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
		c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
	} else {
		updateMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
	}
}

func updateMediaTask(currentJob *job.Job, mediaRequest *UpdateMediaRequest, configs *util.Configs, c *gin.Context, wait bool) {
	currentJob.SetExecutingStateWithValue("Updating media on the DB", mediaRequest.Name)
	err := updateMedia(mediaRequest, configs)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetCompletedState("Media updated on DB")
	if wait {
		c.JSON(http.StatusOK, gin.H{"message": "Media updated on DB"})
	}
}

func updateMedia(mediaRequest *UpdateMediaRequest, configs *util.Configs) error {
	dbPath := filepath.Join(configs.Database.FolderPath, "trackers.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	stm, err := db.Prepare(`
UPDATE
	medias_tracker
SET
    media_type = ?,
	priority = ?,
	status = ?,
	stars = ?,
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
		mediaRequest.MediaType,
		mediaRequest.Priority,
		mediaRequest.Status,
		mediaRequest.Stars,
		mediaRequest.StartedDate,
		mediaRequest.FinishedDroppedDate,
		mediaRequest.Commentary,
		mediaRequest.ReleaseDate,
		mediaRequest.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

type UpdateMediaRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Whether the requester wants to wait for the task to be done before responding
	Name                   string    `json:"name" binding:"required"`
	MediaType              int       `json:"media_type" binding:"required"`
	Priority               int       `json:"priority" binding:"required"`
	Status                 int       `json:"status" binding:"required"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	ReleaseDateStr         string    `json:"release_date" binding:"omitempty,IsValidDate"`
	ReleaseDate            time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (mr *UpdateMediaRequest) GetStartedDateStr() string {
	return mr.StartedDateStr
}

func (mr *UpdateMediaRequest) GetFinishedDroppedDateStr() string {
	return mr.FinishedDroppedDateStr
}

func (mr *UpdateMediaRequest) GetReleaseDateStr() string {
	return mr.ReleaseDateStr
}

func (mr *UpdateMediaRequest) SetStartedDate(startedDate time.Time) {
	mr.StartedDate = startedDate
}

func (mr *UpdateMediaRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	mr.FinishedDroppedDate = finishedDroppedDate
}

func (mr *UpdateMediaRequest) SetReleaseDate(releaseDate time.Time) {
	mr.ReleaseDate = releaseDate
}
