package trackers_test

import (
	"bytes"
	"encoding/json"
	"github.com/diogovalentte/dashboard/api/scraping"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/routes/trackers"
	"github.com/diogovalentte/dashboard/api/util"
)

func TestGetMediaMetadata(t *testing.T) {
	expected := trackers.ScrapedMediaProperties{
		Name:        "Fight Club",
		CoverURL:    "https://m.media-amazon.com/images/M/MV5BODQ0OWJiMzktYjNlYi00MzcwLThlZWMtMzRkYTY4ZDgxNzgxXkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_QL75_UX190_CR0,2,190,281_.jpg",
		ReleaseDate: time.Date(1999, 10, 15, 0, 0, 0, 0, time.UTC),
		Genres:      []string{"Drama"},
		Staff:       []string{"David Fincher", "Chuck Palahniuk", "Jim Uhls", "Brad Pitt", "Edward Norton", "Meat Loaf"},
	}

	// Get media metadata
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		t.Error(err)
		return
	}
	wd, geckodriver, err := scraping.GetWebDriver((*configs).Firefox.BinaryPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer wd.Close()
	defer geckodriver.Release()

	mediaURL := "https://www.imdb.com/title/tt0137523"
	actual, err := trackers.GetMediaMetadata(mediaURL, &wd)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(expected, *actual) {
		t.Errorf("expected: %s, actual: %s", expected, *actual)
		return
	}

	t.Logf("Media scraped: %s", actual.Name)
}

var addMediaRouteTestTable = []*trackers.AddMediaRequest{
	{
		Wait:      true,
		MediaType: 3,
		URL:       "https://www.imdb.com/title/tt1865718/?ref_=fn_al_tt_1",
		Priority:  1,
		Status:    4,
	},
	{
		Wait:                   true,
		URL:                    "https://www.imdb.com/title/tt1586680",
		MediaType:              1,
		Priority:               2,
		Status:                 3,
		Stars:                  5,
		StartedDateStr:         "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
		Commentary:             "Shameless",
	},
	{
		Wait:                   true,
		URL:                    "https://www.imdb.com/title/tt0468569/?ref_=chttp_t_3",
		MediaType:              2,
		Priority:               2,
		Status:                 3,
		Stars:                  5,
		StartedDateStr:         "2023-07-29",
		FinishedDroppedDateStr: "2023-08-12",
		Commentary:             "Batman",
	},
}

func TestAddMediaRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, mediaRequest := range addMediaRouteTestTable {
		// Make request
		requestBody, err := json.Marshal(mediaRequest)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/medias_tracker/add_media", bytes.NewBuffer(requestBody))
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

		// Validate
		expectedMessage := "Media added to DB"
		actualMessage, exists := resMap["message"]
		if !exists {
			t.Error(`Response body has no field "message"`)
			continue
		}

		if http.StatusOK != w.Code {
			if actualMessage == "UNIQUE constraint failed: medias_tracker.name" {
				t.Log("Media already in database")
				continue
			} else {
				t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
				t.Errorf(`expected message: %s, actual message: %s`, expectedMessage, actualMessage)
				continue
			}
		} else {
			if !exists {
				t.Error(`Response body has no field "message"`)
			}
			if actualMessage != expectedMessage {
				t.Errorf(`expected message: %s, actual message: %s`, expectedMessage, actualMessage)
			}

			t.Log(actualMessage)
		}
	}
}

var addMediaManuallyRouteTestTable = []*trackers.MediaProperties{
	{
		Wait:                   true,
		URL:                    "https://www.imdb.com/title/tt9140554/?ref_=hm_top_tt_i_3",
		MediaType:              3,
		Priority:               1,
		Status:                 4,
		Stars:                  3,
		Name:                   "Loki",
		CoverImgURL:            "https://m.media-amazon.com/images/M/MV5BYTY0YTgwZjUtYzJiNy00ZDQ2LWFlZmItZThhMjExMjI5YWQ2XkEyXkFqcGdeQXVyMTM1NjM2ODg1._V1_QL75_UX190_CR0,0,190,281_.jpg",
		Genres:                 []string{},
		Staff:                  []string{},
		ReleaseDateStr:         "2021-06-09",
		StartedDateStr:         "2023-01-01",
		FinishedDroppedDateStr: "2023-01-02",
		Commentary:             "Cool.",
	},
	{
		Wait:                   true,
		URL:                    "https://www.imdb.com/title/tt1517268/",
		MediaType:              1,
		Priority:               2,
		Status:                 3,
		Stars:                  3,
		Name:                   "Barbie",
		CoverImgURL:            "https://m.media-amazon.com/images/M/MV5BNjU3N2QxNzYtMjk1NC00MTc4LTk1NTQtMmUxNTljM2I0NDA5XkEyXkFqcGdeQXVyODE5NzE3OTE@._V1_QL75_UX190_CR0,0,190,281_.jpg",
		ReleaseDateStr:         "2023-07-21",
		StartedDateStr:         "2022-12-01",
		FinishedDroppedDateStr: "2023-01-05",
	},
}

func TestAddMediaManuallyRoute(t *testing.T) {
	router := api.SetupRouter()

	for _, mediaProperties := range addMediaManuallyRouteTestTable {
		// Make request
		requestBody, err := json.Marshal(mediaProperties)
		if err != nil {
			t.Error(err)
			continue
		}

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/trackers/medias_tracker/add_media_manually", bytes.NewBuffer(requestBody))
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

		// Validate
		expectedMessage := "Media added to DB"
		actualMessage, exists := resMap["message"]
		if !exists {
			t.Error(`Response body has no field "message"`)
			continue
		}

		if http.StatusOK != w.Code {
			if actualMessage == "UNIQUE constraint failed: medias_tracker.name" {
				t.Log("Media already in database")
				continue
			} else {
				t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
				t.Errorf(`expected message: %s, actual message: %s`, expectedMessage, actualMessage)
				continue
			}
		} else {
			if !exists {
				t.Error(`Response body has no field "message"`)
			}
			if actualMessage != expectedMessage {
				t.Errorf(`expected message: %s, actual message: %s`, expectedMessage, actualMessage)
			}

			t.Log(actualMessage)
		}
	}
}
