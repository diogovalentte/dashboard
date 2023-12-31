package trackers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/diogovalentte/dashboard/api/job"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/tebeka/selenium"
)

func AddGame(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Add game to Games Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
		return
	}
	jobsList.AddJob(&currentJob)
	currentJob.SetStartingState("Processing game request")

	// Validate request
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("IsValidDate", IsValidDate)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	var gameRequest AddGameRequest
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

	// Get game info from a web store and insert into DB
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if !gameRequest.Wait {
		go addGameTask(&currentJob, nil, configs, &gameRequest)
	} else {
		addGameTask(&currentJob, c, configs, &gameRequest)
	}
}

type AddGameRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Whether the requester wants to wait for the task to be done before responding
	URL                    string    `json:"url" binding:"required,http_url"`
	Priority               int       `json:"priority" binding:"required"`
	Status                 int       `json:"status" binding:"required"`
	PurchasedGamePass      bool      `json:"purchased_or_gamepass" binding:"-"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (gr *AddGameRequest) GetStartedDateStr() string {
	return gr.StartedDateStr
}

func (gr *AddGameRequest) GetFinishedDroppedDateStr() string {
	return gr.FinishedDroppedDateStr
}

func (gr *AddGameRequest) GetReleaseDateStr() string {
	return ""
}

func (gr *AddGameRequest) SetStartedDate(startedDate time.Time) {
	gr.StartedDate = startedDate
}

func (gr *AddGameRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	gr.FinishedDroppedDate = finishedDroppedDate
}

func (gr *AddGameRequest) SetReleaseDate(releaseDate time.Time) {}

func addGameTask(currentJob *job.Job, context *gin.Context, configs *util.Configs, gameRequest *AddGameRequest) {
	// Get webdriver
	currentJob.SetExecutingStateWithValue("Waiting for a WebDriver", gameRequest.URL)
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

	// Scrap game metadata
	currentJob.SetExecutingState("Scraping game data")
	scrapedGameProperties, err := GetGameMetadata(gameRequest.URL, &wd)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Create GameProperties
	gameProperties, err := getGameProperties(gameRequest, scrapedGameProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Insert game into DB
	currentJob.SetExecutingStateWithValue("Adding game to DB", scrapedGameProperties.Name)
	err = insertGameIntoDB(gameProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		if context != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	completeMsg := "Game added to DB"
	currentJob.SetCompletedStateWithValue(completeMsg, gameProperties.Name)
	if context != nil {
		context.JSON(http.StatusOK, gin.H{"message": completeMsg})
	}
}

func GetGameMetadata(gameURL string, wd *selenium.WebDriver) (*ScrapedGameProperties, error) {
	// Get game metadata from a web store (Steam)
	gameURL = strings.SplitN(gameURL, "?", 2)[0]
	steamPrefix := "https://store.steampowered.com/app/"
	if isSteamURL := strings.HasPrefix(gameURL, steamPrefix); !isSteamURL {
		return nil, fmt.Errorf("the game url %s is not a valid Steam url, it should start with: %s", gameURL, steamPrefix)
	}

	// Get the game properties
	if err := (*wd).Get(gameURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s. Error: %s", gameURL, err)
	}

	timeout := 10 * time.Second
	secondAttempt, thirdAttempt := false, false
	for !thirdAttempt {
		err := (*wd).WaitWithTimeout(gameNameCondition, timeout)
		if err != nil {
			timeoutErrorPrefix := "timeout after"
			isTimeoutError := strings.HasPrefix(err.Error(), timeoutErrorPrefix)
			if isTimeoutError {
				// Maybe is an age consent page where we need to submit an age to get the game page
				optionsXPATH := map[string]string{"//select[@id='ageDay']": "29", "//select[@id='ageMonth']": "August", "//select[@id='ageYear']": "1958"} // King's birthday
				for xpath, option := range optionsXPATH {
					err = selectFromDropdown(wd, xpath, option)
					if err != nil {
						return nil, err
					}
				}

				viewPageElem, err := (*wd).FindElement(selenium.ByXPATH, "//a[@id='view_product_page_btn']")
				if err != nil {
					return nil, err
				}
				if err = viewPageElem.Click(); err != nil {
					return nil, err
				}

				if !secondAttempt {
					secondAttempt = true
				} else {
					thirdAttempt = true
				}
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	// Name
	gameNameElem, err := (*wd).FindElement(selenium.ByXPATH, "//div[@id='appHubAppName']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	gameName, err := gameNameElem.Text()
	if err != nil {
		return nil, err
	}

	// Cover URL
	coverURLElem, err := (*wd).FindElement(selenium.ByXPATH, "//img[@class='game_header_image_full']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err
	}

	// Release date
	var releaseDate time.Time
	var secondErr error
	releaseDateElem, err := (*wd).FindElement(selenium.ByXPATH, "//div[@class='release_date']/div[@class='date']")
	if err != nil {
		if err.Error() != "no such element: Unable to locate element: //div[@class='release_date']/div[@class='date']" {
			return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
		}
	} else {
		releaseDateStr, err := releaseDateElem.Text()
		if err != nil {
			return nil, err
		}

		switch releaseDateStr {
		case "To be announced":
		default:
			steamLayout := "2 Jan, 2006"
			releaseDate, err = time.Parse(steamLayout, releaseDateStr)
			if err != nil {
				steamLayout := "Jan 2, 2006"
				releaseDate, secondErr = time.Parse(steamLayout, releaseDateStr)
				if secondErr != nil {
					return nil, err
				}
			}
		}
	}

	// Tags
	var tags []string
	tagsElems, err := (*wd).FindElements(selenium.ByXPATH, "//div[contains(@class, 'glance_tags popular_tags')]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	tags, err = getTextFromDisplayNoneElements(tagsElems, wd)
	if err != nil {
		return nil, err
	}

	// Developers
	var developers []string
	developersElems, err := (*wd).FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Developer:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	developers, err = getTextFromDisplayNoneElements(developersElems, wd)
	if err != nil {
		return nil, err
	}

	// Publishers
	var publishers []string
	publishersElems, err := (*wd).FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Publisher:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	publishers, err = getTextFromDisplayNoneElements(publishersElems, wd)
	if err != nil {
		return nil, err
	}

	scrapedGameProperties := ScrapedGameProperties{
		Name:        gameName,
		CoverURL:    coverURL,
		ReleaseDate: releaseDate,
		Developers:  developers,
		Publishers:  publishers,
		Tags:        tags,
	}

	return &scrapedGameProperties, nil
}

type ScrapedGameProperties struct {
	Name        string
	CoverURL    string
	ReleaseDate time.Time
	Tags        []string
	Developers  []string
	Publishers  []string
}

func gameNameCondition(wd selenium.WebDriver) (bool, error) {
	_, err := wd.FindElement(selenium.ByXPATH, "//div[@id='appHubAppName']")
	if err != nil {
		return false, nil
	}

	return true, nil
}

func selectFromDropdown(wd *selenium.WebDriver, xpath, option string) error {
	elem, err := (*wd).FindElement(selenium.ByXPATH, xpath)
	if err != nil {
		return err
	}
	if err = elem.Click(); err != nil {
		return err
	}
	time.Sleep(2 * time.Second) // Ensure the dropdown options are visible

	if err = elem.SendKeys(option); err != nil {
		return err
	}
	if err = elem.SendKeys(selenium.EnterKey); err != nil {
		return err
	}

	return nil
}

func getTextFromDisplayNoneElements(elems []selenium.WebElement, wd *selenium.WebDriver) ([]string, error) {
	// Some tags have style "display: none", so this work around is needed to get the tag text
	var elemsText []string
	for _, elem := range elems {
		script := "return arguments[0].textContent"
		var args []interface{}
		args = append(args, elem)
		res, err := (*wd).ExecuteScript(script, args)
		if err != nil {
			return nil, err
		}

		elemText := res.(string)
		elemText = strings.TrimSpace(elemText)
		elemsText = append(elemsText, elemText)
	}

	return elemsText, nil
}

func getGameProperties(gr *AddGameRequest, sgp *ScrapedGameProperties) (*GameProperties, error) {
	coverImg, err := util.GetImageFromURL(sgp.CoverURL)
	if err != nil {
		return nil, err
	}

	tags := strings.Join(sgp.Tags, ",")
	developers := strings.Join(sgp.Developers, ",")
	publishers := strings.Join(sgp.Publishers, ",")

	return &GameProperties{
		Name:                sgp.Name,
		URL:                 gr.URL,
		CoverImg:            coverImg,
		ReleaseDate:         sgp.ReleaseDate,
		TagsStr:             tags,
		DevelopersStr:       developers,
		PublishersStr:       publishers,
		Priority:            gr.Priority,
		Status:              gr.Status,
		Stars:               gr.Stars,
		PurchasedOrGamePass: gr.PurchasedGamePass,
		StartedDate:         gr.StartedDate,
		FinishedDroppedDate: gr.FinishedDroppedDate,
		Commentary:          gr.Commentary,
	}, nil
}

type GameProperties struct {
	Wait                   bool   `json:"wait" binding:"-"` // Whether the requester wants to wait for the task to be done before responding
	URL                    string `json:"url" binding:"required"`
	Priority               int    `json:"priority" binding:"required"`
	Status                 int    `json:"status" binding:"required"`
	Stars                  int    `json:"stars" binding:"omitempty,gte=0,lte=5"`
	PurchasedOrGamePass    bool   `json:"purchased_or_gamepass" binding:"-"`
	Name                   string `json:"name" binding:"required"`
	CoverImgURL            string `json:"cover_img_url" binding:"required"`
	CoverImg               []byte
	Tags                   []string `json:"tags" binding:"-"`
	TagsStr                string
	Developers             []string `json:"developers" binding:"-"`
	DevelopersStr          string
	Publishers             []string `json:"publishers" binding:"-"`
	PublishersStr          string
	ReleaseDateStr         string    `json:"release_date" binding:"omitempty,IsValidDate"`
	ReleaseDate            time.Time `binding:"-"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (gr *GameProperties) GetStartedDateStr() string {
	return gr.StartedDateStr
}

func (gr *GameProperties) GetFinishedDroppedDateStr() string {
	return gr.FinishedDroppedDateStr
}

func (gr *GameProperties) GetReleaseDateStr() string {
	return gr.ReleaseDateStr
}

func (gr *GameProperties) SetStartedDate(startedDate time.Time) {
	gr.StartedDate = startedDate
}

func (gr *GameProperties) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	gr.FinishedDroppedDate = finishedDroppedDate
}

func (gr *GameProperties) SetReleaseDate(releaseDate time.Time) {
	gr.ReleaseDate = releaseDate
}

func insertGameIntoDB(gp *GameProperties) error {
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
INSERT INTO games_tracker (
  url, name, cover_img, release_date, tags, developers, publishers, priority,
  status, stars, purchased_or_gamepass, started_date, finished_dropped_date, commentary
)
VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
  `)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(
		gp.URL,
		gp.Name,
		gp.CoverImg,
		gp.ReleaseDate,
		gp.TagsStr,
		gp.DevelopersStr,
		gp.PublishersStr,
		gp.Priority,
		gp.Status,
		gp.Stars,
		gp.PurchasedOrGamePass,
		gp.StartedDate,
		gp.FinishedDroppedDate,
		gp.Commentary,
	)
	if err != nil {
		return err
	}

	return nil
}

func AddGameManually(c *gin.Context) {
	// Create job
	currentJob := job.Job{
		Task:      "Add game to Games Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
		return
	}
	jobsList.AddJob(&currentJob)
	currentJob.SetStartingState("Processing game request")

	// Validate request
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("IsValidDate", IsValidDate)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	var gameProperties GameProperties
	if err := c.ShouldBindJSON(&gameProperties); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation: %s", err.Error())
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := SetStructDateFields(&gameProperties)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Convert properties
	coverImg, err := util.GetImageFromURL(gameProperties.CoverImgURL)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	gameProperties.CoverImg = coverImg

	gameProperties.TagsStr = strings.Join(gameProperties.Tags, ",")
	gameProperties.DevelopersStr = strings.Join(gameProperties.Developers, ",")
	gameProperties.PublishersStr = strings.Join(gameProperties.Publishers, ",")

	// Insert game into DB
	currentJob.SetExecutingStateWithValue("Adding game to DB", gameProperties.Name)
	if !gameProperties.Wait {
		go func(currentJob *job.Job, gameProperties *GameProperties) {
			err := insertGameIntoDB(gameProperties)
			if err != nil {
				currentJob.SetFailedState(err)
				return
			}

			currentJob.SetCompletedState("Game added to DB")
		}(&currentJob, &gameProperties)
	} else {
		err = insertGameIntoDB(&gameProperties)
		if err != nil {
			currentJob.SetFailedState(err)
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game added to DB"})
	}
}
