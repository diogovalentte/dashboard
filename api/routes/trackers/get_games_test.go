package trackers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/trackers"
)

func TestGetAllGamesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/games_tracker/get_all_games", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
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
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
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
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
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
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
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
	}
	router.ServeHTTP(w, req)

	var res getGamesResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
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
