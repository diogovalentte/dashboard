package notion

import (
	"context"
	"strings"
	"time"

	"github.com/tebeka/selenium"

	"github.com/diogovalentte/dashboard/api/util"
	"github.com/diogovalentte/notionapi"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GamesTrackerRoutes(group *gin.RouterGroup) {
	games_tracker_group := group.Group("/games_tracker")
	{
		games_tracker_group.POST("/add_game", AddGame)
	}
}

func MediasTrackerRoutes(group *gin.RouterGroup) {
	medias_tracker_group := group.Group("/medias_tracker")
	{
		medias_tracker_group.POST("/add_media", AddMedia)
	}
}

func IsValidDate(fl validator.FieldLevel) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, fl.Field().String())

	return err == nil
}

func GetNotionClient() (*notionapi.Client, error) {
	configs, err := util.GetConfigs()
	if err != nil {
		return nil, err
	}

	return notionapi.NewClient(notionapi.Token(configs.Notion.Token)), nil
}

type StructDateFields interface {
	GetStartedDateStr() string
	GetFinishedDroppedDateStr() string
	SetStartedDate(time.Time)
	SetFinishedDroppedDate(time.Time)
}

func SetStructDateFields(input StructDateFields) error {
	layout := "2006-01-02"

	startedDateStr := input.GetStartedDateStr()
	if startedDateStr != "" {
		startedDate, err := time.Parse(layout, startedDateStr)
		if err != nil {
			return err
		}
		input.SetStartedDate(startedDate)
	}
	finishedDroppedDateStr := input.GetFinishedDroppedDateStr()
	if finishedDroppedDateStr != "" {
		finishedDroppedDate, err := time.Parse(layout, finishedDroppedDateStr)
		if err != nil {
			return err
		}
		input.SetFinishedDroppedDate(finishedDroppedDate)
	}

	return nil
}

func transformIntoMultiSelect(array *[]string) *[]notionapi.Option {
	var options []notionapi.Option
	for _, elem := range *array {
		elem := strings.ReplaceAll(elem, ",", "")
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
	client, err := GetNotionClient()
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

func getTextFromElements(elems []selenium.WebElement) ([]string, error) {
	var elemsText []string
	for _, elem := range elems {
		text, err := elem.Text()
		if err != nil {
			return []string{}, err
		}
		elemsText = append(elemsText, text)
	}

	return elemsText, nil
}
