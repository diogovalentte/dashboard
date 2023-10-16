package trackers_test

import (
	"bytes"
	"encoding/json"
	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/trackers"
	"net/http"
	"net/http/httptest"
	"testing"
)

var updateGameRouteTestTable = []*trackers.UpdateGameRequest{
	{
		Wait:                   true,
		Name:                   "Terraria",
		Priority:               3,
		Status:                 3,
		PurchasedGamePass:      false,
		Stars:                  3,
		StartedDateStr:         "2023-01-01",
		FinishedDroppedDateStr: "2023-01-02",
		ReleaseDateStr:         "2023-01-03",
		Commentary:             "Not my type.",
	},
	{
		Wait:                   true,
		Name:                   "Red Dead Redemption 2",
		Priority:               1,
		Status:                 5,
		PurchasedGamePass:      true,
		Stars:                  5,
		StartedDateStr:         "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
		ReleaseDateStr:         "2023-08-09",
		Commentary:             "One of the best games of all time.",
	},
	{
		Wait:                   true,
		Name:                   "Remnant II",
		Priority:               2,
		Status:                 3,
		PurchasedGamePass:      false,
		Stars:                  5,
		StartedDateStr:         "2023-07-29",
		FinishedDroppedDateStr: "2023-08-12",
		ReleaseDateStr:         "2023-02-13",
		Commentary:             "The biggest surprise of 2023.",
	},
}

func TestUpdateGameRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, gameRequest := range updateGameRouteTestTable {
		requestBody, err := json.Marshal(gameRequest)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/games_tracker/update_game", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(w, req)

		var resMap map[string]string
		jsonBytes := w.Body.Bytes()
		if err := json.Unmarshal(jsonBytes, &resMap); err != nil {
			t.Error(err)
			continue
		}

		if http.StatusOK != w.Code {
			t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
		}

		expectedMessage := "Game updated on DB"
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
