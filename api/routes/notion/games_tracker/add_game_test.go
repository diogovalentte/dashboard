package games_tracker_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/notion/games_tracker"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
)

func startGeckoDriverServer(GeckoDriverPath string, Port int) (*scraping.GeckoDriverServer, error) {
	gds := scraping.NewGeckoDriverServer(GeckoDriverPath, Port)

	go gds.Start()

	err := gds.Wait()
	if err != nil {
		return nil, err
	}

	return gds, nil
}

func stopGeckoDriverServer(gds *scraping.GeckoDriverServer) error {
	return gds.Stop()
}

func TestMain(m *testing.M) {
	configs, err := util.GetConfigsWithoutDefaults("../../../../configs")
	if err != nil {
		panic(err)
	}

	gds, err := startGeckoDriverServer(configs.GeckoDriver.BinaryPath, configs.GeckoDriver.Port)
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	err = stopGeckoDriverServer(gds)
	if err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}

func TestGetGameMetadata(t *testing.T) {
	configs, err := util.GetConfigsWithoutDefaults("../../../../configs")
	if err != nil {
		panic(err)
	}

	expected := games_tracker.ScrapedGameProperties{
		Name:        "Red Dead Redemption 2",
		CoverURL:    "https://cdn.akamai.steamstatic.com/steam/apps/1174180/header.jpg?t=1671485009",
		ReleaseDate: time.Date(2019, 12, 5, 0, 0, 0, 0, time.UTC),
		Developers:  []string{"Rockstar Games"},
		Publishers:  []string{"Rockstar Games"},
		Tags:        []string{"Open World", "Story Rich", "Western", "Adventure", "Action", "Multiplayer", "Realistic", "Singleplayer", "Shooter", "Atmospheric", "Horses", "Beautiful", "Third-Person Shooter", "Mature", "Great Soundtrack", "Third Person", "Sandbox", "Gore", "First-Person", "FPS"},
	}

	gameURL := "https://store.steampowered.com/app/1174180/Red_Dead_Redemption_2"
	actual, err := games_tracker.GetGameMetadata(gameURL, configs.Firefox.BinaryPath, configs.GeckoDriver.Port)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, *actual) {
		t.Errorf("expected: %s, actual: %s", expected, *actual)
	}
}

var addGameRouteTestTable = []*games_tracker.GameRequest{
	{
		URL:                    "https://store.steampowered.com/app/105600/Terraria/",
		Priority:               "Low",
		Status:                 "Dropped",
		PurchasedGamePass:      false,
		Stars:                  3,
		StartedPlayingStr:      "2023-01-01",
		FinishedDroppedDateStr: "2023-01-02",
		Commentary:             "Not my type.",
	},
	{
		URL:                    "https://store.steampowered.com/app/1174180/Red_Dead_Redemption_2/?l=brazilian",
		Priority:               "High",
		Status:                 "Done",
		PurchasedGamePass:      true,
		Stars:                  5,
		StartedPlayingStr:      "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
		Commentary:             "One of the best games of all time.",
	},
	{
		URL:                    "https://store.steampowered.com/app/1282100/Remnant_II/",
		Priority:               "Medium",
		Status:                 "Done",
		PurchasedGamePass:      false,
		Stars:                  5,
		StartedPlayingStr:      "2023-07-29",
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

		pageURL, exists := resMap["page_url"]
		if !exists {
			t.Error(`Response body has no field "page_url"`)
		}

		t.Log(pageURL)
	}
}
