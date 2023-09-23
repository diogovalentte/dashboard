package games_tracker

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

// State
func AddGame(c *gin.Context) {
	// Create the job
	currentJob := job.Job{
		Task:      "Add game to Games Tracker database",
		State:     "Starting",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	currentJob.SetStartingState("Processing game request")

	jobsList, ok := c.MustGet("JobsList").(*job.Jobs)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create the task's job"})
	}
	jobsList.AddJob(&currentJob)

	// Start the task
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("IsValidDate", util.IsValidDate)
	}

	var gameRequest GameRequest
	if err := c.ShouldBindJSON(&gameRequest); err != nil {
		err = fmt.Errorf("invalid JSON fields, refer to the API documentation")
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := setDateFields(&gameRequest)
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configs, err := util.GetConfigs()
	if err != nil {
		currentJob.SetFailedState(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get game info from a web store and create the game notion page
	go func() {
		currentJob.SetExecutingStateWithValue("Scraping game data", gameRequest.URL)

		scrapedGameProperties, err, responseError := GetGameMetadata(gameRequest.URL, (*configs).Firefox.BinaryPath, (*configs).GeckoDriver.Port)
		if err != nil {
			if responseError {
				currentJob.SetFailedState(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			} else {
				err = fmt.Errorf("couldn't get the game properties from site")
				currentJob.SetFailedState(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		currentJob.SetExecutingStateWithValue("Creating game page", scrapedGameProperties.Name)

		gameProperties := mergeToGameProperties(&gameRequest, scrapedGameProperties)
		_, notionPageURL, err, responseError := createGamePage(gameProperties, (*configs).Notion.GameTracker.GamesDBID)
		if err != nil {
			if responseError {
				currentJob.SetFailedState(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			} else {
				err = fmt.Errorf("couldn't create the game page")
				currentJob.SetFailedState(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		currentJob.SetCompletedStateWithValue("Game page created", notionPageURL)
	}()
	c.JSON(http.StatusOK, gin.H{"message": "Job created with success"})
}

func gameNameCondition(wd selenium.WebDriver) (bool, error) {
	_, err := wd.FindElement(selenium.ByXPATH, "//div[@id='appHubAppName']")
	if err != nil {
		return false, nil
	}

	return true, nil
}

type GameRequest struct {
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

func setDateFields(gr *GameRequest) error {
	layout := "2006-01-02"

	if gr.StartedDateStr != "" {
		startedDate, err := time.Parse(layout, gr.StartedDateStr)
		if err != nil {
			return err
		}
		gr.StartedDate = startedDate
	}
	if gr.FinishedDroppedDateStr != "" {
		finishedDroppedDate, err := time.Parse(layout, gr.FinishedDroppedDateStr)
		if err != nil {
			return err
		}
		gr.FinishedDroppedDate = finishedDroppedDate
	}

	return nil
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

func GetGameMetadata(gameURL, firefoxPath string, geckoDriverServerPort int) (*ScrapedGameProperties, error, bool) {
	// Gets game metadata from a web store (Steam)
	gameURL = strings.SplitN(gameURL, "?", 2)[0]
	steamPrefix := "https://store.steampowered.com/app/"
	if isSteamURL := strings.HasPrefix(gameURL, steamPrefix); !isSteamURL {
		return nil, fmt.Errorf("the game url %s is not a valid Steam url, it should start with: %s", gameURL, steamPrefix), true
	}

	wd, err := scraping.GetWebDriver(firefoxPath, geckoDriverServerPort)
	if err != nil {
		return nil, err, false
	}
	defer wd.Close()

	// Get the game properties
	if err := wd.Get(gameURL); err != nil {
		return nil, fmt.Errorf("could not get the page with URL: %s", gameURL), true
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
						return nil, err, false
					}
				}

				viewPageElem, err := wd.FindElement(selenium.ByXPATH, "//a[@id='view_product_page_btn']")
				if err != nil {
					return nil, err, false
				}
				if err = viewPageElem.Click(); err != nil {
					return nil, err, false
				}

				if !secondAttempt {
					secondAttempt = true
				} else {
					thirdAttempt = true
				}
			} else {
				return nil, err, false
			}
		} else {
			break
		}
	}

	// Name
	gameNameElem, err := wd.FindElement(selenium.ByXPATH, "//div[@id='appHubAppName']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	gameName, err := gameNameElem.Text()
	if err != nil {
		return nil, err, false
	}

	// Cover URL
	coverURLElem, err := wd.FindElement(selenium.ByXPATH, "//img[@class='game_header_image_full']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err, false
	}

	// Release date
	releaseDateElem, err := wd.FindElement(selenium.ByXPATH, "//div[@class='release_date']/div[@class='date']")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	releaseDateStr, err := releaseDateElem.Text()
	if err != nil {
		return nil, err, false
	}

	steamLayout := "2 Jan, 2006"
	releaseDate, err := time.Parse(steamLayout, releaseDateStr)
	if err != nil {
		return nil, err, false
	}

	// Tags
	var tags []string
	tagsElems, err := wd.FindElements(selenium.ByXPATH, "//div[contains(@class, 'glance_tags popular_tags')]/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	tags, err = getTextFromElements(tagsElems, &wd)
	if err != nil {
		return nil, err, false
	}

	// Developers
	var developers []string
	developersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Developer:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	developers, err = getTextFromElements(developersElems, &wd)
	if err != nil {
		return nil, err, false
	}

	// Publishers
	var publishers []string
	publishersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Publisher:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, fmt.Errorf("couldn't find an element in the page: %s", err), false
	}
	publishers, err = getTextFromElements(publishersElems, &wd)
	if err != nil {
		return nil, err, false
	}

	gameProperties := ScrapedGameProperties{
		Name:        gameName,
		CoverURL:    coverURL,
		ReleaseDate: releaseDate,
		Developers:  developers,
		Publishers:  publishers,
		Tags:        tags,
	}

	return &gameProperties, nil, false
}

type ScrapedGameProperties struct {
	Name        string
	CoverURL    string
	ReleaseDate time.Time
	Tags        []string
	Developers  []string
	Publishers  []string
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

func getTextFromElements(elems []selenium.WebElement, wd *selenium.WebDriver) ([]string, error) {
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

func createGamePage(gp *GameProperties, DB_ID string) (notionapi.ObjectID, string, error, bool) {
	pageCreateRequest := getPageRequest(gp, DB_ID)

	client, err := util.GetNotionClient()
	if err != nil {
		return "", "", err, false
	}

	page, err := client.Page.Create(context.Background(), pageCreateRequest)
	if err != nil {
		return "", "", err, false
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
				return "", "", fmt.Errorf("couldn't add commentary to the game page because of the error: %s. Also tried to archive the page, but an error occuried: %s", err, archivePageErr), false
			}
			return "", "", fmt.Errorf("couldn't add commentary to the game page because of the error: %s", err), false
		}
	}

	return page.ID, page.URL, nil, false
}

func getPageRequest(gp *GameProperties, DB_ID string) *notionapi.PageCreateRequest {
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
	releaseDate := notionapi.Date(gp.ReleaseDate)
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
		"Release date": notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &releaseDate,
			},
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

func transformIntoMultiSelect(array *[]string) *[]notionapi.Option {
	var options []notionapi.Option
	for _, elem := range *array {
		option := notionapi.Option{
			Name: elem,
		}
		options = append(options, option)
	}

	return &options
}

func getStarsEmojis(number int) string {
	emoji := "⭐"
	return strings.Repeat(emoji, number)
}

func ArchivePage(pageId string) (bool, error) {
	// Archive a page, returns wheter the page is archived or not
	client, err := util.GetNotionClient()
	if err != nil {
		return false, err
	}

	page, err := client.Page.Update(context.Background(), notionapi.PageID(pageId), &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{},
		Archived:   true,
	})
	if err != nil {
		return false, err
	}

	return page.Archived, nil
}
