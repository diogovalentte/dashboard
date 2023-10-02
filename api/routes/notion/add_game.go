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

func AddGame(c *gin.Context) {
	// Create and add job
	currentJob := job.Job{
		Task:      "Add game to Games Tracker database",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	currentJob.SetStartingState("Processing game request")

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "couldn't create the task's job"})
	}
	jobsList.AddJob(&currentJob)

	// Validate request
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("IsValidDate", IsValidDate)
	}

	var gameRequest GameRequest
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

	// Get game info from a web store and create the game notion page
	configs, err := util.GetConfigs()
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Get game info from a web store and create the game notion page
	if !gameRequest.Wait {
		go addGameTask(&currentJob, &gameRequest, configs, c, gameRequest.Wait)
		c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
	} else {
		addGameTask(&currentJob, &gameRequest, configs, c, gameRequest.Wait)
	}
}

type GameRequest struct {
	Wait                   bool      `json:"wait" binding:"-"` // Wether the requester wants to wait for the task to be done before responding
	URL                    string    `json:"url" binding:"required,http_url"`
	Priority               string    `json:"priority" binding:"required"`
	Status                 string    `json:"status" binding:"required"`
	PurchasedGamePass      bool      `json:"purchased_or_gamepass" binding:"-"`
	Stars                  int       `json:"stars" binding:"omitempty,gte=0,lte=5"`
	StartedDateStr         string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedDate            time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func (gr *GameRequest) GetStartedDateStr() string {
	return gr.StartedDateStr
}

func (gr *GameRequest) GetFinishedDroppedDateStr() string {
	return gr.FinishedDroppedDateStr
}

func (gr *GameRequest) SetStartedDate(startedDate time.Time) {
	gr.StartedDate = startedDate
}

func (gr *GameRequest) SetFinishedDroppedDate(finishedDroppedDate time.Time) {
	gr.FinishedDroppedDate = finishedDroppedDate
}

func addGameTask(currentJob *job.Job, gameRequest *GameRequest, configs *util.Configs, c *gin.Context, wait bool) {
	currentJob.SetExecutingStateWithValue("Scraping game data", gameRequest.URL)

	scrapedGameProperties, err := GetGameMetadata(gameRequest.URL, (*configs).Firefox.BinaryPath, (*configs).GeckoDriver.Port)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetExecutingStateWithValue("Creating game page", scrapedGameProperties.Name)

	gameProperties := mergeToGameProperties(gameRequest, scrapedGameProperties)
	_, notionPageURL, err := createGamePage(gameProperties, (*configs).Notion.GamesTracker.DBID)
	if err != nil {
		currentJob.SetFailedState(err)
		if wait {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	currentJob.SetCompletedStateWithValue("Game page created", notionPageURL)

	if wait {
		c.JSON(http.StatusOK, gin.H{"message": "Game page created with success"})
	}
}

func GetGameMetadata(gameURL, firefoxPath string, geckoDriverServerPort int) (*ScrapedGameProperties, error) {
	// Gets game metadata from a web store (Steam)
	gameURL = strings.SplitN(gameURL, "?", 2)[0]
	steamPrefix := "https://store.steampowered.com/app/"
	if isSteamURL := strings.HasPrefix(gameURL, steamPrefix); !isSteamURL {
		return nil, fmt.Errorf("the game url %s is not a valid Steam url, it should start with: %s", gameURL, steamPrefix)
	}

	wd, err := scraping.GetWebDriver(firefoxPath, geckoDriverServerPort)
	if err != nil {
		return nil, err
	}
	defer wd.Close()

	// Get the game properties
	if err := wd.Get(gameURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s. Error: %s", gameURL, err)
	}

	timeout := 10 * time.Second
	secondAttempt, thirdAttempt := false, false
	for !thirdAttempt {
		err = wd.WaitWithTimeout(gameNameCondition, timeout)
		if err != nil {
			timeoutErrorPrefix := "timeout after"
			isTimeoutError := strings.HasPrefix(err.Error(), timeoutErrorPrefix)
			if isTimeoutError {
				// Maybe is an age consent page where we need to submit an age to get the game page
				optionsXPATH := map[string]string{"//select[@id='ageDay']": "29", "//select[@id='ageMonth']": "August", "//select[@id='ageYear']": "1958"} // King's birthday
				for xpath, option := range optionsXPATH {
					err = selectFromDropdown(&wd, xpath, option)
					if err != nil {
						return nil, err
					}
				}

				viewPageElem, err := wd.FindElement(selenium.ByXPATH, "//a[@id='view_product_page_btn']")
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
	gameNameElem, err := wd.FindElement(selenium.ByXPATH, "//div[@id='appHubAppName']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	gameName, err := gameNameElem.Text()
	if err != nil {
		return nil, err
	}

	// Cover URL
	coverURLElem, err := wd.FindElement(selenium.ByXPATH, "//img[@class='game_header_image_full']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err
	}

	// Release date
	releaseDateElem, err := wd.FindElement(selenium.ByXPATH, "//div[@class='release_date']/div[@class='date']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	releaseDateStr, err := releaseDateElem.Text()
	if err != nil {
		return nil, err
	}

	var releaseDate time.Time
	switch releaseDateStr {
	case "To be announced":
	default:
		steamLayout := "2 Jan, 2006"
		releaseDate, err = time.Parse(steamLayout, releaseDateStr)
		if err != nil {
			return nil, err
		}
	}

	// Tags
	var tags []string
	tagsElems, err := wd.FindElements(selenium.ByXPATH, "//div[contains(@class, 'glance_tags popular_tags')]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	tags, err = getTextFromDisplayNoneElements(tagsElems, &wd)
	if err != nil {
		return nil, err
	}

	// Developers
	var developers []string
	developersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Developer:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	developers, err = getTextFromDisplayNoneElements(developersElems, &wd)
	if err != nil {
		return nil, err
	}

	// Publishers
	var publishers []string
	publishersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Publisher:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err)
	}
	publishers, err = getTextFromDisplayNoneElements(publishersElems, &wd)
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

func mergeToGameProperties(gr *GameRequest, sgp *ScrapedGameProperties) *GameProperties {
	return &GameProperties{
		Name:                sgp.Name,
		URL:                 gr.URL,
		CoverURL:            sgp.CoverURL,
		ReleaseDate:         sgp.ReleaseDate,
		Tags:                sgp.Tags,
		Developers:          sgp.Developers,
		Publishers:          sgp.Publishers,
		Priority:            gr.Priority,
		Status:              gr.Status,
		Stars:               gr.Stars,
		PurchasedOrGamePass: gr.PurchasedGamePass,
		StartedDate:         gr.StartedDate,
		FinishedDroppedDate: gr.FinishedDroppedDate,
		Commentary:          gr.Commentary,
	}
}

type GameProperties struct {
	Name                string
	URL                 string
	CoverURL            string
	ReleaseDate         time.Time
	Tags                []string
	Developers          []string
	Publishers          []string
	Priority            string
	Status              string
	Stars               int
	PurchasedOrGamePass bool
	StartedDate         time.Time
	FinishedDroppedDate time.Time
	Commentary          string
}

func createGamePage(gp *GameProperties, DB_ID string) (notionapi.ObjectID, string, error) {
	pageCreateRequest := getGamePageRequest(gp, DB_ID)

	client, err := GetNotionClient()
	if err != nil {
		return "", "", err
	}

	page, err := client.Page.Create(context.Background(), pageCreateRequest)
	if err != nil {
		return "", "", err
	}

	// Add commentary
	if gp.Commentary != "" {
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
									Content: gp.Commentary,
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
				return "", "", fmt.Errorf("couldn't add commentary to the game page because of the error: %s. Also tried to archive the page, but an error occuried: %s", err, archivePageErr)
			}
			return "", "", fmt.Errorf("couldn't add commentary to the game page because of the error: %s", err)
		}
	}

	return page.ID, page.URL, nil
}

func getGamePageRequest(gp *GameProperties, DB_ID string) *notionapi.PageCreateRequest {
	validProperties := getValidGameProperties(gp)

	pageCreateRequest := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(DB_ID),
		},
		Cover: &notionapi.Image{
			Type:     "external",
			External: &notionapi.FileObject{URL: gp.CoverURL},
		},
		Properties: *validProperties,
	}

	return pageCreateRequest
}

func getValidGameProperties(gp *GameProperties) *notionapi.Properties {
	tags := transformIntoMultiSelect(&gp.Tags)
	developers := transformIntoMultiSelect(&gp.Developers)
	publishers := transformIntoMultiSelect(&gp.Publishers)

	gameProperties := notionapi.Properties{
		"Name": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{Text: &notionapi.Text{Content: gp.Name}},
			},
		},
		"Link": notionapi.URLProperty{
			URL: gp.URL,
		},
		"Tags": notionapi.MultiSelectProperty{
			MultiSelect: *tags,
		},
		"Developers": notionapi.MultiSelectProperty{
			MultiSelect: *developers,
		},
		"Publishers": notionapi.MultiSelectProperty{
			MultiSelect: *publishers,
		},
		"Priority": notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: gp.Priority,
			},
		},
		"Status": notionapi.StatusProperty{
			Status: notionapi.Option{
				Name: gp.Status,
			},
		},
		"Purchased/GamePass?": notionapi.CheckboxProperty{
			Checkbox: gp.PurchasedOrGamePass,
		},
	}

	// Release date (some games don't have)
	var zeroTime time.Time
	if !gp.ReleaseDate.Equal(zeroTime) {
		releaseDate := notionapi.Date(gp.ReleaseDate)
		gameProperties["Release date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &releaseDate,
			},
		}
	}

	// Optional properties
	if gp.Stars != 0 {
		stars := getStarsEmojis(gp.Stars)
		gameProperties["Stars"] = notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: stars,
			},
		}
	}

	if !gp.StartedDate.IsZero() {
		startedDate := notionapi.Date(gp.StartedDate)
		gameProperties["Started date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &startedDate,
			},
		}
	}

	if !gp.FinishedDroppedDate.IsZero() {
		finishedDroppedDate := notionapi.Date(gp.FinishedDroppedDate)
		gameProperties["Finished/Dropped date"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &finishedDroppedDate,
			},
		}
	}

	return &gameProperties
}
