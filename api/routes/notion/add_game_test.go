package notion_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/notion"
	"github.com/diogovalentte/dashboard/api/util"
)

func TestGetGameMetadata(t *testing.T) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		panic(err)
	}

	expected := notion.ScrapedGameProperties{
		Name:        "Red Dead Redemption 2",
		CoverURL:    "https://cdn.akamai.steamstatic.com/steam/apps/1174180/header.jpg?t=1695140956",
		ReleaseDate: time.Date(2019, 12, 5, 0, 0, 0, 0, time.UTC),
		Developers:  []string{"Rockstar Games"},
		Publishers:  []string{"Rockstar Games"},
		Tags:        []string{"Open World", "Story Rich", "Western", "Adventure", "Action", "Multiplayer", "Realistic", "Singleplayer", "Shooter", "Atmospheric", "Horses", "Beautiful", "Third-Person Shooter", "Mature", "Great Soundtrack", "Third Person", "Sandbox", "Gore", "First-Person", "FPS"},
	}

	gameURL := "https://store.steampowered.com/app/1174180/Red_Dead_Redemption_2"
	actual, err, _ := notion.GetGameMetadata(gameURL, configs.Firefox.BinaryPath, configs.GeckoDriver.Port)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, *actual) {
		t.Errorf("expected: %s, actual: %s", expected, *actual)
	}
}

var addGameRouteTestTable = []*notion.GameRequest{
	{
		Wait:                   true,
		URL:                    "https://store.steampowered.com/app/105600/Terraria/",
		Priority:               "Low",
		Status:                 "Playing",
		PurchasedGamePass:      false,
		Stars:                  3,
		StartedDateStr:         "2023-01-01",
		FinishedDroppedDateStr: "2023-01-02",
		Commentary:             "Not my type.",
	},
	{
		Wait:                   true,
		URL:                    "https://store.steampowered.com/app/1174180/Red_Dead_Redemption_2/?l=brazilian",
		Priority:               "High",
		Status:                 "Finished",
		PurchasedGamePass:      true,
		Stars:                  5,
		StartedDateStr:         "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
		Commentary:             "One of the best games of all time.",
	},
	{
		Wait:                   true,
		URL:                    "https://store.steampowered.com/app/1282100/Remnant_II/",
		Priority:               "Medium",
		Status:                 "Playing",
		PurchasedGamePass:      false,
		Stars:                  5,
		StartedDateStr:         "2023-07-29",
		FinishedDroppedDateStr: "2023-08-12",
		Commentary:             "The biggest surprise of 2023.",
	},
}

func TestAddGameRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, gameRequest := range addGameRouteTestTable {
		requestBody, err := json.Marshal(gameRequest)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/notion/games_tracker/add_game", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Error(err)
		}
		router.ServeHTTP(w, req)

		var resMap map[string]string
		jsonBytes := w.Body.Bytes()
		if err := json.Unmarshal(jsonBytes, &resMap); err != nil {
			t.Error(err)
		}

		if http.StatusOK != w.Code {
			t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
		}

		expectedMessage := "Game page created with success"
		actualMessage, exists := resMap["message"]
		if !exists {
			t.Error(`Response body has no field "message"`)
		}
		if actualMessage != expectedMessage {
			t.Errorf(`expected message: %s, actual message: %s`, expectedMessage, actualMessage)
		}

		t.Log(actualMessage)
	}
}
