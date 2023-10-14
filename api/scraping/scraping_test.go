package scraping

import (
	"github.com/diogovalentte/dashboard/api/util"
	"testing"
)

func TestGeckoDriverPoolLifeCycle(t *testing.T) {
	configs, err := util.GetConfigsWithoutDefaults("../../configs")
	if err != nil {
		t.Error(err)
		return
	}

	// Start pool
	size := 2
	t.Log("Starting new geckodriver pool")
	pool, err := NewGeckoDriverPool(configs.GeckoDriver.BinaryPath, size)
	if err != nil {
		t.Error(err)
		return
	}

	// Get geckodriver instances
	gdSlice := make([]*GeckoDriverServer, size)
	for i := 0; i < size; i++ {
		t.Logf("Getting geckodriver number %d", i)
		gds, err := pool.WaitGet()
		if err != nil {
			t.Error(err)
			return
		}
		gdSlice[i] = gds
	}

	// Release the instances
	for i := 0; i < size; i++ {
		t.Logf("Releasing geckodriver number %d", i)

		gdSlice[i].Release()
	}

	// Stop pool instances
	t.Log("Stopping pool")
	err = pool.StopAll()
	if err != nil {
		t.Error(err)
		return
	}
}

var testIsPortAvailableTable = [4]int{
	8000, 8080, 8888, 5555,
}

func TestIsPortAvailable(t *testing.T) {
	for _, port := range testIsPortAvailableTable {
		t.Logf("Testing port: %d", port)
		_, err := isPortAvailable(port)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
