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

func TestGetMediaMetadata(t *testing.T) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		panic(err)
	}

	expected := notion.ScrapedMediaProperties{
		Name:        "Fight Club",
		CoverURL:    "https://m.media-amazon.com/images/M/MV5BODQ0OWJiMzktYjNlYi00MzcwLThlZWMtMzRkYTY4ZDgxNzgxXkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_QL75_UX190_CR0,2,190,281_.jpg",
		ReleaseDate: time.Date(1999, 10, 15, 0, 0, 0, 0, time.UTC),
		Genres:      []string{"Drama"},
		Staff:       []string{"David Fincher", "Chuck Palahniuk", "Jim Uhls", "Brad Pitt", "Edward Norton", "Meat Loaf"},
	}

	mediaURL := "https://www.imdb.com/title/tt0137523"
	actual, err, _ := notion.GetMediaMetadata(mediaURL, configs.Firefox.BinaryPath, configs.GeckoDriver.Port)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, *actual) {
		t.Errorf("expected: %s, actual: %s", expected, *actual)
	}
}

var addMediaRouteTestTable = []*notion.MediaRequest{
	// {
	// 	Wait:      true,
	// 	MediaType: "Series",
	// 	URL:       "https://www.imdb.com/title/tt1865718/?ref_=fn_al_tt_1",
	// 	Priority:  "High",
	// 	Status:    "Finished",
	// },
	// {
	// 	Wait:                   true,
	// 	URL:                    "https://www.imdb.com/title/tt1586680",
	// 	MediaType:              "Series",
	// 	Priority:               "High",
	// 	Status:                 "To be released",
	// 	Stars:                  5,
	// 	StartedDateStr:         "2022-12-01",
	// 	FinishedDroppedDateStr: "2023-01-05",
	// 	Commentary:             "Shameless",
	// },
	{
		Wait:                   true,
		URL:                    "https://www.imdb.com/title/tt0468569/?ref_=chttp_t_3",
		MediaType:              "Movie",
		Priority:               "Medium",
		Status:                 "Watching/Reading",
		Stars:                  5,
		StartedDateStr:         "2023-07-29",
		FinishedDroppedDateStr: "2023-08-12",
		Commentary:             "Batman",
	},
}

func TestAddMediaRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, mediaRequest := range addMediaRouteTestTable {
		requestBody, err := json.Marshal(mediaRequest)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/notion/medias_tracker/add_media", bytes.NewBuffer(requestBody))
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

		expectedMessage := "Media page created with success"
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
