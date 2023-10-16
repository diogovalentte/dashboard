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

var updateMediaRouteTestTable = []*trackers.UpdateMediaRequest{
	{
		Wait:                   true,
		Name:                   "Gravity Falls",
		MediaType:              1,
		Priority:               1,
		Status:                 1,
		Stars:                  0,
		StartedDateStr:         "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
		Commentary:             "Gravity Up",
	},
	{
		Wait:                   true,
		Name:                   "Shameless",
		MediaType:              1,
		Priority:               1,
		Status:                 1,
		Stars:                  0,
		StartedDateStr:         "2022-02-01",
		FinishedDroppedDateStr: "2023-12-05",
		Commentary:             "Shamefull",
	},
	{
		Wait:                   true,
		Name:                   "The Dark Knight",
		MediaType:              1,
		Priority:               1,
		Status:                 1,
		Stars:                  0,
		StartedDateStr:         "2022-01-15",
		FinishedDroppedDateStr: "2023-08-23",
		Commentary:             "The Shiny Knight",
	},
}

func TestUpdateMediaRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, mediaRequest := range updateMediaRouteTestTable {
		requestBody, err := json.Marshal(mediaRequest)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/medias_tracker/update_media", bytes.NewBuffer(requestBody))
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

		expectedMessage := "Media updated on DB"
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
