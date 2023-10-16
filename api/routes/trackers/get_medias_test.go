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

var getMediaRouteTestTable = []trackers.GetMediaRequest{
	{
		Name: "Gravity Falls",
	},
	{
		Name: "Shameless",
	},
	{
		Name: "The Dark Knight",
	},
}

func TestGetMediaRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, mediaRequest := range getMediaRouteTestTable {
		requestBody, err := json.Marshal(mediaRequest)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/medias_tracker/get_media", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(w, req)

		var res getMediaResponse
		jsonBytes := w.Body.Bytes()
		if err := json.Unmarshal(jsonBytes, &res); err != nil {
			t.Error(err)
			continue
		}

		if http.StatusOK != w.Code {
			t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
			t.Errorf("message: %s", res.Message)
			continue
		}

		media := res.Media
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

func TestGetAllMediasRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/medias_tracker/get_all_medias", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getMediasResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	medias := res.Medias
	for _, media := range medias {
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

func TestGetToBeReleasedMediasRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/medias_tracker/get_to_be_released_medias", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getMediasResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	medias := res.Medias
	for _, media := range medias {
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

func TestGetNotStartedMediasRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/medias_tracker/get_not_started_medias", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getMediasResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	medias := res.Medias
	for _, media := range medias {
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

func TestGetFinishedMediasRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/medias_tracker/get_finished_medias", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getMediasResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	medias := res.Medias
	for _, media := range medias {
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

func TestGetDroppedMediasRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/trackers/medias_tracker/get_dropped_medias", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(w, req)

	var res getMediasResponse
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	medias := res.Medias
	for _, media := range medias {
		t.Log(fmt.Sprintf("Media: %s", media.Name))
	}
}

type getMediasResponse struct {
	Medias []trackers.MediaProperties `json:"medias"`
}

type getMediaResponse struct {
	Media   trackers.MediaProperties `json:"media"`
	Message string                   `json:"message"`
}
