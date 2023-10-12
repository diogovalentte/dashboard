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
