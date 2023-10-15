package trackers

import (
	"time"

	"github.com/tebeka/selenium"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GamesTrackerRoutes(group *gin.RouterGroup) {
	games_tracker_group := group.Group("/games_tracker")
	{
		games_tracker_group.POST("/add_game", AddGame)
		games_tracker_group.GET("/get_all_games", GetAllGames)
		games_tracker_group.GET("/get_playing_games", GetPlayingGames)
		games_tracker_group.GET("/get_to_be_released_games", GetToBeReleasedGames)
		games_tracker_group.GET("/get_not_started_games", GetNotStartedGames)
		games_tracker_group.GET("/get_finished_games", GetFinishedGames)
		games_tracker_group.GET("/get_dropped_games", GetDroppedGames)
	}
}

func MediasTrackerRoutes(group *gin.RouterGroup) {
	medias_tracker_group := group.Group("/medias_tracker")
	{
		medias_tracker_group.POST("/add_media", AddMedia)
		medias_tracker_group.GET("/get_all_medias", GetAllMedias)
		medias_tracker_group.GET("/get_watching_reading_medias", GetWatchingReadingMedias)
		medias_tracker_group.GET("/get_to_be_released_medias", GetToBeReleasedMedias)
		medias_tracker_group.GET("/get_not_started_medias", GetNotStartedMedias)
		medias_tracker_group.GET("/get_finished_medias", GetFinishedMedias)
		medias_tracker_group.GET("/get_dropped_medias", GetDroppedMedias)
	}
}

func IsValidDate(fl validator.FieldLevel) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, fl.Field().String())

	return err == nil
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
