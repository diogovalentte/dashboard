package trackers_test

import (
	"fmt"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
	"os"
	"testing"
)

func setup() (*scraping.GeckoDriverPool, error) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		return nil, err
	}

	// Start the GeckoDriver server
	pool, err := scraping.NewGeckoDriverPool(configs.GeckoDriver.BinaryPath, 1)
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
