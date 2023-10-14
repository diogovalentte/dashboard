package system_test

import (
	"encoding/json"
	"fmt"
	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetGeckoDriverInstancesRoute(t *testing.T) {
	router := api.SetupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v1/system/get_geckodrivers", nil)
	if err != nil {
		t.Error(err)
		return
	}
	router.ServeHTTP(w, req)

	var res getGeckoDrivers
	jsonBytes := w.Body.Bytes()
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		t.Error(err)
		return
	}

	if http.StatusOK != w.Code {
		t.Errorf("expected status code: %d, actual status code: %d", http.StatusOK, w.Code)
	}

	addresses := res.Addresses
	for _, addr := range addresses {
		t.Log(fmt.Sprintf("Addr: %s", addr))
	}
}

type getGeckoDrivers struct {
	Addresses []string `json:"addresses"`
}

func setup() (*scraping.GeckoDriverPool, error) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		return nil, err
	}

	// Start the GeckoDriver server
	pool, err := scraping.NewGeckoDriverPool(configs.GeckoDriver.BinaryPath, 3)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func teardown(pool *scraping.GeckoDriverPool) error {
	err := pool.StopAll()
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	pool, err := setup()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	result := m.Run()

	err = teardown(pool)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(result)
}
