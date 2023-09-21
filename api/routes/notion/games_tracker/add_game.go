package games_tracker

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"github.com/diogovalentte/notionapi"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/tebeka/selenium"
)

func AddGame(c *gin.Context) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("IsValidDate", util.IsValidDate)
	}

	var gameRequest GameRequest
	if err := c.ShouldBindJSON(&gameRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON fields, refer to the API documentation"})
		return
	}

	err := setDateFields(&gameRequest)
	if err != nil {
		panic(err)
	}

	configs, err := util.GetConfigs()
	if err != nil {
		panic(err)
	}

	scrapedGameProperties, err := GetGameMetadata(gameRequest.URL, (*configs).Firefox.BinaryPath, (*configs).GeckoDriver.Port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get the game properties from site"})
	}

	gameProperties := mergeToGameProperties(&gameRequest, scrapedGameProperties)
	_, notionPageURL, err := createGamePage(gameProperties, (*configs).Notion.GameTracker.GamesDBID)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, map[string]string{"page_url": notionPageURL})
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
	StartedPlayingStr      string    `json:"started_date" binding:"omitempty,IsValidDate"`
	StartedPlaying         time.Time `binding:"-"`
	FinishedDroppedDateStr string    `json:"finished_dropped_date" binding:"omitempty,IsValidDate"`
	FinishedDroppedDate    time.Time `binding:"-"`
	Commentary             string    `json:"commentary" binding:"-"`
}

func setDateFields(gr *GameRequest) error {
	layout := "2006-01-02"

	if gr.StartedPlayingStr != "" {
		startedPlaying, err := time.Parse(layout, gr.StartedPlayingStr)
		if err != nil {
			return err
		}
		gr.StartedPlaying = startedPlaying
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
		StartedPlaying:      gr.StartedPlaying,
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
	StartedPlaying      time.Time
	FinishedDroppedDate time.Time
	Commentary          string
}

func GetGameMetadata(gameURL, firefoxPath string, geckoDriverServerPort int) (*ScrapedGameProperties, error) {
	// Gets game metadata from a web store (Steam)
	gameURL = strings.SplitN(gameURL, "?", 2)[0]

	wd, err := scraping.GetWebDriver(firefoxPath, geckoDriverServerPort)
	if err != nil {
		return nil, err
	}
	defer wd.Close()

	// Get the game properties
	if err := wd.Get(gameURL); err != nil {
		return nil, err
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
		return nil, err
	}
	gameName, err := gameNameElem.Text()
	if err != nil {
		return nil, err
	}

	// Cover URL
	coverURLElem, err := wd.FindElement(selenium.ByXPATH, "//img[@class='game_header_image_full']")
	if err != nil {
		return nil, err
	}
	coverURL, err := coverURLElem.GetAttribute("src")
	if err != nil {
		return nil, err
	}

	// Release date
	releaseDateElem, err := wd.FindElement(selenium.ByXPATH, "//div[@class='release_date']/div[@class='date']")
	if err != nil {
		return nil, err
	}
	releaseDateStr, err := releaseDateElem.Text()
	if err != nil {
		return nil, err
	}

	steamLayout := "2 Jan, 2006"
	releaseDate, err := time.Parse(steamLayout, releaseDateStr)
	if err != nil {
		return nil, err
	}

	// Tags
	var tags []string
	tagsElems, err := wd.FindElements(selenium.ByXPATH, "//div[contains(@class, 'glance_tags popular_tags')]/a")
	if err != nil {
		return nil, err
	}
	tags, err = getTextFromElements(tagsElems, &wd)
	if err != nil {
		return nil, err
	}

	// Developers
	var developers []string
	developersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Developer:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, err
	}
	developers, err = getTextFromElements(developersElems, &wd)
	if err != nil {
		return nil, err
	}

	// Publishers
	var publishers []string
	publishersElems, err := wd.FindElements(selenium.ByXPATH, "//div[@class='dev_row']/div[contains(@class, 'subtitle')][text()='Publisher:']/../div[@class='summary column']/a")
	if err != nil {
		return nil, err
	}
	publishers, err = getTextFromElements(publishersElems, &wd)
	if err != nil {
		return nil, err
	}

	gameProperties := ScrapedGameProperties{
		Name:        gameName,
		CoverURL:    coverURL,
		ReleaseDate: releaseDate,
		Developers:  developers,
		Publishers:  publishers,
		Tags:        tags,
	}

	return &gameProperties, nil
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

func createGamePage(gp *GameProperties, DB_ID string) (notionapi.ObjectID, string, error) {
	pageCreateRequest, err := getPageRequest(gp, DB_ID)
	if err != nil {
		return "", "", err
	}

	client, err := util.GetNotionClient()
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

func getPageRequest(gp *GameProperties, DB_ID string) (*notionapi.PageCreateRequest, error) {
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

	return pageCreateRequest, nil
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

	if !gp.StartedPlaying.IsZero() {
		startedPlaying := notionapi.Date(gp.StartedPlaying)
		gameProperties["Started playing"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: &startedPlaying,
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
