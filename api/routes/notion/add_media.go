package notion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/diogovalentte/notionapi"
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
	currentJob.SetStartingState("Processing media request")

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create the task's job"})
	}
	jobsList.AddJob(&currentJob)

	// Validate request
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("IsValidDate", IsValidDate)
	}

	var mediaRequest MediaRequest
	if err := c.ShouldBindJSON(&mediaRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := SetStructDateFields(&mediaRequest)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get media info from a media site and create the media notion page
	configs, err := util.GetConfigs()
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get media info from a web store and create the media notion page
	if !mediaRequest.Wait {
		go addMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
		c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
	} else {
		addMediaTask(&currentJob, &mediaRequest, configs, c, mediaRequest.Wait)
	}
}

type MediaRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Wether the requester wants to wait for the task to be done before responding
	URL                    string    `json:"url" binding:"required,http_url"`
	MediaType              string    `json:"type" binding:"required"`
	Priority               string    `json:"priority" binding:"required"`
	Status                 string    `json:"status" binding:"required"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (mr *MediaRequest) GetStartedDateStr() string {
	return mr.StartedDateStr
}

func (mr *MediaRequest) GetFinishedDroppedDateStr() string {
	return mr.FinishedDroppedDateStr
}

func (mr *MediaRequest) SetStartedDate(startedDate time.Time) {
	mr.StartedDate = startedDate
}

func (mr *MediaRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	mr.FinishedDroppedDate = finishedDroppedDate
}

func addMediaTask(currentJob *job.Job, mediaRequest *MediaRequest, configs *util.Configs, c *gin.Context, wait bool) {
	currentJob.SetExecutingStateWithValue("Scraping media data", mediaRequest.URL)

	scrapedMediaProperties, err, setErrorAsStateDescription := GetMediaMetadata(mediaRequest.URL, (*configs).Firefox.BinaryPath, (*configs).GeckoDriver.Port)
	if err != nil {
		if setErrorAsStateDescription {
			currentJob.SetFailedState(err)
			if wait {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
			return
		} else {
			err = fmt.Errorf("couldn't get the media properties from site")
			currentJob.SetFailedState(err)
			if wait {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
			return
		}
	}

	currentJob.SetExecutingStateWithValue("Creating media page", scrapedMediaProperties.Name)

	mediaProperties := mergeToMediaProperties(mediaRequest, scrapedMediaProperties)
	_, notionPageURL, err, setErrorAsStateDescription := createMediaPage(mediaProperties, (*configs).Notion.MediasTracker.DBID)
	if err != nil {
		if setErrorAsStateDescription {
			currentJob.SetFailedState(err)
			if wait {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
			return
		} else {
			err = fmt.Errorf("couldn't create the media page")
			currentJob.SetFailedState(err)
			if wait {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
			return
		}
	}

	currentJob.SetCompletedStateWithValue("Media page created", notionPageURL)

	if wait {
		c.JSON(http.StatusOK, gin.H{"message": "Media page created with success"})
	}
}

func GetMediaMetadata(mediaURL, firefoxPath string, geckoDriverServerPort int) (*ScrapedMediaProperties, error, bool) {
	// Gets media metadata from a media site (IMDB)
	mediaURL = strings.SplitN(mediaURL, "?", 2)[0]
	IMDB_Prefix := "https://www.imdb.com/title/"
	if isIMDB_URL := strings.HasPrefix(mediaURL, IMDB_Prefix); !isIMDB_URL {
		return nil, fmt.Errorf("the media url %s is not a valid IMDB url, it should start with: %s", mediaURL, IMDB_Prefix), true
	}

	wd, err := scraping.GetWebDriver(firefoxPath, geckoDriverServerPort)
	if err != nil {
		return nil, err, false
	}
	defer wd.Close()

	// Get the media properties
	if err := wd.Get(mediaURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s", mediaURL), true
	}

	timeout := 10 * time.Second
	err = wd.WaitWithTimeout(mediaNameCondition, timeout)
	if err != nil {
		return nil, fmt.Errorf("timeout while waiting for page to load"), true
	}

	// Name
	mediaNameElem, err := wd.FindElement(selenium.ByXPATH, "//h1[@data-testid='hero__pageTitle']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	mediaName, err := mediaNameElem.Text()
	if err != nil {
		return nil, err, false
	}

	// Cover URL
	coverURLElem, err := wd.FindElement(selenium.ByXPATH, "//*[contains(@class, 'ipc-media--poster-l')]//img[@class='ipc-image']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err, false
	}

	// Release date
	releaseDateElem, err := wd.FindElement(selenium.ByXPATH, "//a[text()='Release date']/..//ul/li/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	releaseDateStr, err := releaseDateElem.Text()
	if err != nil {
		return nil, err, false
	}

	releaseDateStr = strings.SplitN(releaseDateStr, " (", 2)[0]
	steamLayout := "January 2, 2006"
	releaseDate, err := time.Parse(steamLayout, releaseDateStr)
	if err != nil {
		return nil, err, false
	}

	// Genres
	genreElems, err := wd.FindElements(selenium.ByXPATH, "(//div[@class='ipc-chip-list__scroller'])[1]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	genres, err := getTextFromElements(genreElems)
	if err != nil {
		return nil, err, false
	}

	// Staff
	staffElems, err := wd.FindElements(selenium.ByXPATH, "(//ul[@class='ipc-metadata-list ipc-metadata-list--dividers-all title-pc-list ipc-metadata-list--baseAlt'])[1]/li//li/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	staff, err := getTextFromElements(staffElems)
	if err != nil {
		return nil, err, false
	}

	scrapedMediaProperties := ScrapedMediaProperties{
		Name:        mediaName,
		CoverURL:    coverURL,
		ReleaseDate: releaseDate,
		Genres:      genres,
		Staff:       staff,
	}

	return &scrapedMediaProperties, nil, false
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

func mergeToMediaProperties(mr *MediaRequest, smp *ScrapedMediaProperties) *MediaProperties {
	return &MediaProperties{
		Name:                smp.Name,
		URL:                 mr.URL,
		MediaType:           mr.MediaType,
		CoverURL:            smp.CoverURL,
		ReleaseDate:         smp.ReleaseDate,
		Genres:              smp.Genres,
		Staff:               smp.Staff,
		Priority:            mr.Priority,
		Status:              mr.Status,
		Stars:               mr.Stars,
		StartedDate:         mr.StartedDate,
		FinishedDroppedDate: mr.FinishedDroppedDate,
		Commentary:          mr.Commentary,
	}
}

type MediaProperties struct {
	Name                string
	URL                 string
	MediaType           string
	CoverURL            string
	ReleaseDate         time.Time
	Genres              []string
	Staff               []string
	Priority            string
	Status              string
	Stars               int
	StartedDate         time.Time
	FinishedDroppedDate time.Time
	Commentary          string
}

func createMediaPage(mp *MediaProperties, DB_ID string) (notionapi.ObjectID, string, error, bool) {
	pageCreateRequest := getMediaPageRequest(mp, DB_ID)

	client, err := GetNotionClient()
	if err != nil {
		return "", "", err, false
	}

	page, err := client.Page.Create(context.Background(), pageCreateRequest)
	if err != nil {
		return "", "", err, false
	}

	// Add commentary
	if mp.Commentary != "" {
		_, err = client.Block.AppendChildren(context.Background(), notionapi.BlockID(page.ID), &notionapi.AppendBlockChildrenRequest{
			Children: []notionapi.Block{
				notionapi.QuoteBlock{
					BasicBlock: notionapi.BasicBlock{
						Type:   notionapi.BlockType("quote"),
						Object: notionapi.ObjectType("block"),
					},
					Quote: notionapi.Quote{
						RichText: []notionapi.RichText{
							{
								Type: notionapi.ObjectType("text"),
								Text: &notionapi.Text{
									Content: mp.Commentary,
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			_, archivePageErr := ArchivePage(page.ID.String())
			if archivePageErr != nil {
				return "", "", fmt.Errorf("couldn't add commentary to the media page because of the error: %s. Also tried to archive the page, but an error occuried: %s", err, archivePageErr), false
			}
			return "", "", fmt.Errorf("couldn't add commentary to the media page because of the error: %s", err), false
		}
	}

	return page.ID, page.URL, nil, false
}

func getMediaPageRequest(mp *MediaProperties, DB_ID string) *notionapi.PageCreateRequest {
	validProperties := getValidMediaProperties(mp)

	pageCreateRequest := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(DB_ID),
		},
		Cover: &notionapi.Image{
			Type:     "external",
			External: &notionapi.FileObject{URL: mp.CoverURL},
		},
		Properties: *validProperties,
	}

	return pageCreateRequest
}

func getValidMediaProperties(mp *MediaProperties) *notionapi.Properties {
	releaseDate := notionapi.Date(mp.ReleaseDate)
	genres := transformIntoMultiSelect(&mp.Genres)
	staff := transformIntoMultiSelect(&mp.Staff)

	mediaProperties := notionapi.Properties{
		"Name": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{Text: &notionapi.Text{Content: mp.Name}},
			},
		},
		"Link": notionapi.URLProperty{
			URL: mp.URL,
		},
		"Type": notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: mp.MediaType,
			},
		},
		"Release date": notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &releaseDate,
			},
		},
		"Genres": notionapi.MultiSelectProperty{
			MultiSelect: *genres,
		},
		"Staff": notionapi.MultiSelectProperty{
			MultiSelect: *staff,
		},
		"Priority": notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: mp.Priority,
			},
		},
		"Status": notionapi.StatusProperty{
			Status: notionapi.Option{
				Name: mp.Status,
			},
		},
	}

	// Optional properties
	if mp.Stars != 0 {
		stars := getStarsEmojis(mp.Stars)
		mediaProperties["Stars"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: stars,
			},
		}
	}

	if !mp.StartedDate.IsZero() {
		startedDate := notionapi.Date(mp.StartedDate)
		mediaProperties["Started date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &startedDate,
			},
		}
	}

	if !mp.FinishedDroppedDate.IsZero() {
		finishedDroppedDate := notionapi.Date(mp.FinishedDroppedDate)
		mediaProperties["Finished/Dropped date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &finishedDroppedDate,
			},
		}
	}

	return &mediaProperties
}
