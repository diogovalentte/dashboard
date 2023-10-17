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

	// Get media info from a media site and create the media trackers page
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get media info from a web site and create the media trackers page
	if !mediaRequest.Wait {
		go addMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
		c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
	} else {
		addMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
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

func addMediaTask(currentJob *job.Job, mediaRequest *AddMediaRequest, configs *util.Configs, c *gin.Context, wait bool) {
	scrapedMediaProperties, err := GetMediaMetadata(mediaRequest.URL, (*configs).Firefox.BinaryPath, currentJob)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetExecutingStateWithValue("Creating media page", scrapedMediaProperties.Name)

	mediaProperties, err := getMediaProperties(mediaRequest, scrapedMediaProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	err = insertMediaToDB(mediaProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetCompletedState("Media inserted into DB")

	if wait {
		c.JSON(http.StatusOK, gin.H{"message": "Media inserted into DB"})
	}
}

func GetMediaMetadata(mediaURL, firefoxPath string, job *job.Job) (*ScrapedMediaProperties, error) {
	// Gets media metadata from a media site (IMDB)
	mediaURL = strings.SplitN(mediaURL, "?", 2)[0]
	IMDBPrefix := "https://www.imdb.com/title/"
	if isIMDB_URL := strings.HasPrefix(mediaURL, IMDBPrefix); !isIMDB_URL {
		return nil, fmt.Errorf("the media url %s is not a valid IMDB url, it should start with: %s", mediaURL, IMDBPrefix)
	}

	job.SetExecutingStateWithValue("Waiting for a WebDriver", mediaURL)
	wd, geckodriver, err := scraping.GetWebDriver(firefoxPath)
	if err != nil {
		return nil, err
	}
	defer wd.Close()
	defer geckodriver.Release()
	job.SetExecutingState("Scraping media data")

	// Get the media properties
	if err := wd.Get(mediaURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s. Error: %s", mediaURL, err)
	}

	timeout := 10 * time.Second
	err = wd.WaitWithTimeout(mediaNameCondition, timeout)
	if err != nil {
		return nil, fmt.Errorf("timeout while waiting for page to load")
	}

	// Name
	mediaNameElem, err := wd.FindElement(selenium.ByXPATH, "//h1[@data-testid='hero__pageTitle']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	mediaName, err := mediaNameElem.Text()
	if err != nil {
		return nil, err
	}

	// Cover URL
	coverURLElem, err := wd.FindElement(selenium.ByXPATH, "//*[contains(@class, 'ipc-media--poster-l')]//img[@class='ipc-image']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err
	}

	// Release date
	releaseDateElem, err := wd.FindElement(selenium.ByXPATH, "//a[text()='Release date']/..//ul/li/a")
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
	genreElems, err := wd.FindElements(selenium.ByXPATH, "(//div[@class='ipc-chip-list__scroller'])[1]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	genres, err := getTextFromElements(genreElems)
	if err != nil {
		return nil, err
	}

	// Staff
	staffElems, err := wd.FindElements(selenium.ByXPATH, "(//ul[@class='ipc-metadata-list ipc-metadata-list--dividers-all title-pc-list ipc-metadata-list--baseAlt'])[1]/li//li/a")
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
	Name                string
	URL                 string
	MediaType           int
	CoverImg            []byte
	ReleaseDate         time.Time
	GenresStr           string
	Genres              []string
	StaffStr            string
	Staff               []string
	Priority            int
	Status              int
	Stars               int
	StartedDate         time.Time
	FinishedDroppedDate time.Time
	Commentary          string
}

func insertMediaToDB(mp *MediaProperties) error {
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
