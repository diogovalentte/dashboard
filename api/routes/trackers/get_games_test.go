package trackers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/trackers"
)

var getGameRouteTestTable = []trackers.GetGameRequest{
	{
		Name: "Terraria",
	},
	{
		Name: "Red Dead Redemption 2",
	},
	{
		Name: "Remnant II",
	},
}

func TestGetGameRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, gameRequest := range getGameRouteTestTable {
		requestBody, err := json.Marshal(gameRequest)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/games_tracker/get_game", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(w, req)

		var res getGameResponse
		jsonBytes := w.Body.Bytes()
		if err := json.Unmarshal(jsonBytes, &res); err != nil {
			t.Error(err)
			continue
		}

		if http.StatusOK != w.Code {
			t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
		}

		game := res.Game
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

func TestGetAllGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_all_games", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	games := res.Games
	for _, game := range games {
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

func TestToBeReleasedGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_to_be_released_games", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	games := res.Games
	for _, game := range games {
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

func TestNotStartedGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_not_started_games", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	games := res.Games
	for _, game := range games {
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

func TestFinishedGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_finished_games", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	games := res.Games
	for _, game := range games {
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

func TestDroppedGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_dropped_games", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	games := res.Games
	for _, game := range games {
		t.Log(fmt.Sprintf("Game: %s", game.Name))
	}
}

type getGamesResponse struct {
	Games []trackers.GameProperties `json:"games"`
}

type getGameResponse struct {
	Game trackers.GameProperties `json:"game"`
}
