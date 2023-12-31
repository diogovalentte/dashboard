package trackers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/tebeka/selenium"
)

func AddMedia(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Add media to Medias Tracker database",
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

	var mediaRequest AddMediaRequest
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

	// Get media info from a medias site and insert into DB
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if !mediaRequest.Wait {
		go addMediaTask(&currentJob, nil, configs, &mediaRequest)
	} else {
		addMediaTask(&currentJob, c, configs, &mediaRequest)
	}
}

type AddMediaRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Wether the requester wants to wait for the task to be done before responding
	URL                    string    `json:"url" binding:"required,http_url"`
	MediaType              int       `json:"type" binding:"required"`
	Priority               int       `json:"priority" binding:"required"`
	Status                 int       `json:"status" binding:"required"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (mr *AddMediaRequest) GetStartedDateStr() string {
	return mr.StartedDateStr
}

func (mr *AddMediaRequest) GetFinishedDroppedDateStr() string {
	return mr.FinishedDroppedDateStr
}

func (mr *AddMediaRequest) GetReleaseDateStr() string {
	return ""
}

func (mr *AddMediaRequest) SetStartedDate(startedDate time.Time) {
	mr.StartedDate = startedDate
}

func (mr *AddMediaRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	mr.FinishedDroppedDate = finishedDroppedDate
}

func (mr *AddMediaRequest) SetReleaseDate(releaseDate time.Time) {}

func addMediaTask(currentJob *job.Job, context *gin.Context, configs *util.Configs, mediaRequest *AddMediaRequest) {
	// Get webdriver
	currentJob.SetExecutingStateWithValue("Waiting for a WebDriver", mediaRequest.URL)
	wd, geckodriver, err := scraping.GetWebDriver((*configs).Firefox.BinaryPath)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	defer wd.Close()
	defer geckodriver.Release()

	// Scrap media metadata
	currentJob.SetExecutingState("Scraping media data")
	scrapedMediaProperties, err := GetMediaMetadata(mediaRequest.URL, &wd)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Create MediaProperties
	mediaProperties, err := getMediaProperties(mediaRequest, scrapedMediaProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Insert media into DB
	currentJob.SetExecutingStateWithValue("Adding media to DB", scrapedMediaProperties.Name)
	err = insertMediaIntoDB(mediaProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	completeMsg := "Media added to DB"
	currentJob.SetCompletedStateWithValue(completeMsg, mediaProperties.Name)
	if context != nil {
		context.JSON(http.StatusOK, gin.H{"message": completeMsg})
	}
}

func GetMediaMetadata(mediaURL string, wd *selenium.WebDriver) (*ScrapedMediaProperties, error) {
	// Get media metadata from a media site (IMDB)
	mediaURL = strings.SplitN(mediaURL, "?", 2)[0]
	IMDBPrefix := "https://www.imdb.com/title/"
	if isIMDB_URL := strings.HasPrefix(mediaURL, IMDBPrefix); !isIMDB_URL {
		return nil, fmt.Errorf("the media url %s is not a valid IMDB url, it should start with: %s", mediaURL, IMDBPrefix)
	}

	// Get the media properties
	if err := (*wd).Get(mediaURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s. Error: %s", mediaURL, err)
	}

	timeout := 10 * time.Second
	err := (*wd).WaitWithTimeout(mediaNameCondition, timeout)
	if err != nil {
		return nil, fmt.Errorf("timeout while waiting for page to load")
	}

	// Name
	mediaNameElem, err := (*wd).FindElement(selenium.ByXPATH, "//h1[@data-testid='hero__pageTitle']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	mediaName, err := mediaNameElem.Text()
	if err != nil {
		return nil, err
	}

	// Cover URL
	coverURLElem, err := (*wd).FindElement(selenium.ByXPATH, "//*[contains(@class, 'ipc-media--poster-l')]//img[@class='ipc-image']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err
	}

	// Release date
	releaseDateElem, err := (*wd).FindElement(selenium.ByXPATH, "//a[text()='Release date']/..//ul/li/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	releaseDateStr, err := releaseDateElem.Text()
	if err != nil {
		return nil, err
	}

	releaseDateStr = strings.SplitN(releaseDateStr, " (", 2)[0]
	steamLayout := "January 2, 2006"
	releaseDate, err := time.Parse(steamLayout, releaseDateStr)
	if err != nil {
		return nil, err
	}

	// Genres
	genreElems, err := (*wd).FindElements(selenium.ByXPATH, "(//div[@class='ipc-chip-list__scroller'])[1]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	genres, err := getTextFromElements(genreElems)
	if err != nil {
		return nil, err
	}

	// Staff
	staffElems, err := (*wd).FindElements(selenium.ByXPATH, "(//ul[@class='ipc-metadata-list ipc-metadata-list--dividers-all title-pc-list ipc-metadata-list--baseAlt'])[1]/li//li/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	staff, err := getTextFromElements(staffElems)
	if err != nil {
		return nil, err
	}

	scrapedMediaProperties := ScrapedMediaProperties{
		Name:        mediaName,
		CoverURL:    coverURL,
		ReleaseDate: releaseDate,
		Genres:      genres,
		Staff:       staff,
	}

	return &scrapedMediaProperties, nil
}

type ScrapedMediaProperties struct {
	Name        string
	CoverURL    string
	ReleaseDate time.Time
	Genres      []string
	Staff       []string
}

func mediaNameCondition(wd selenium.WebDriver) (bool, error) {
	_, err := wd.FindElement(selenium.ByXPATH, "//h1[@data-testid='hero__pageTitle']")
	if err != nil {
		return false, nil
	}

	return true, nil
}

func getMediaProperties(mr *AddMediaRequest, smp *ScrapedMediaProperties) (*MediaProperties, error) {
	coverImg, err := util.GetImageFromURL(smp.CoverURL)
	if err != nil {
		return nil, err
	}

	genres := strings.Join(smp.Genres, ",")
	staff := strings.Join(smp.Staff, ",")

	return &MediaProperties{
		Name:                smp.Name,
		URL:                 mr.URL,
		MediaType:           mr.MediaType,
		CoverImg:            coverImg,
		ReleaseDate:         smp.ReleaseDate,
		GenresStr:           genres,
		StaffStr:            staff,
		Priority:            mr.Priority,
		Status:              mr.Status,
		Stars:               mr.Stars,
		StartedDate:         mr.StartedDate,
		FinishedDroppedDate: mr.FinishedDroppedDate,
		Commentary:          mr.Commentary,
	}, nil
}

type MediaProperties struct {
	Wait                   bool   `json:"wait" binding:"-"` // Whether the requester wants to wait for the task to be done before responding
	URL                    string `json:"url" binding:"required"`
	Priority               int    `json:"priority" binding:"required"`
	Status                 int    `json:"status" binding:"required"`
	Stars                  int    `json:"stars" binding:"omitempty,gte=0,lte=5"`
	MediaType              int    `json:"media_type" binding:"required"`
	Name                   string `json:"name" binding:"required"`
	CoverImgURL            string `json:"cover_img_url" binding:"required"`
	CoverImg               []byte
	Genres                 []string `json:"genres" binding:"-"`
	GenresStr              string
	Staff                  []string `json:"staff" binding:"-"`
	StaffStr               string
	ReleaseDateStr         string    `json:"release_date" binding:"omitempty,IsValidDate"`
	ReleaseDate            time.Time `binding:"-"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string
}

func (gr *MediaProperties) GetStartedDateStr() string {
	return gr.StartedDateStr
}

func (gr *MediaProperties) GetFinishedDroppedDateStr() string {
	return gr.FinishedDroppedDateStr
}

func (gr *MediaProperties) GetReleaseDateStr() string {
	return gr.ReleaseDateStr
}

func (gr *MediaProperties) SetStartedDate(startedDate time.Time) {
	gr.StartedDate = startedDate
}

func (gr *MediaProperties) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	gr.FinishedDroppedDate = finishedDroppedDate
}

func (gr *MediaProperties) SetReleaseDate(releaseDate time.Time) {
	gr.ReleaseDate = releaseDate
}

func insertMediaIntoDB(mp *MediaProperties) error {
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
INSERT INTO medias_tracker (
  url, name, media_type, cover_img, release_date, genres, staff, priority,
  status, stars, started_date, finished_dropped_date, commentary
)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
  `)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(
		mp.URL,
		mp.Name,
		mp.MediaType,
		mp.CoverImg,
		mp.ReleaseDate,
		mp.GenresStr,
		mp.StaffStr,
		mp.Priority,
		mp.Status,
		mp.Stars,
		mp.StartedDate,
		mp.FinishedDroppedDate,
		mp.Commentary,
	)
	if err != nil {
		return err
	}

	return nil
}

func AddMediaManually(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Add media to Meidas Tracker database",
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

	var mediaProperties MediaProperties
	if err := c.ShouldBindJSON(&mediaProperties); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation: %s", err.Error())
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := SetStructDateFields(&mediaProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Convert properties
	coverImg, err := util.GetImageFromURL(mediaProperties.CoverImgURL)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	mediaProperties.CoverImg = coverImg

	mediaProperties.GenresStr = strings.Join(mediaProperties.Genres, ",")
	mediaProperties.StaffStr = strings.Join(mediaProperties.Staff, ",")

	// Insert media into DB
	currentJob.SetExecutingStateWithValue("Adding media to DB", mediaProperties.Name)
	if !mediaProperties.Wait {
		go func(currentJob *job.Job, mediaProperties *MediaProperties) {
			err := insertMediaIntoDB(mediaProperties)
			if err != nil {
				currentJob.SetFailedState(err)
				return
			}

			currentJob.SetCompletedState("Media added to DB")
		}(&currentJob, &mediaProperties)
	} else {
		err = insertMediaIntoDB(&mediaProperties)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Media added to DB"})
	}
}
