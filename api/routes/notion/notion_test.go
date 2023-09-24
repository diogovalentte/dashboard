package notion_test

import (
	"os"
	"testing"

	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
)

func startGeckoDriverServer(GeckoDriverPath string, Port int) (*scraping.GeckoDriverServer, error) {
	gds := scraping.NewGeckoDriverServer(GeckoDriverPath, Port)

	go gds.Start()

	err := gds.Wait()
	if err != nil {
		return nil, err
	}

	return gds, nil
}

func stopGeckoDriverServer(gds *scraping.GeckoDriverServer) error {
	return gds.Stop()
}

func TestMain(m *testing.M) {
	configs, err := util.GetConfigsWithoutDefaults("../../../configs")
	if err != nil {
		panic(err)
	}

	gds, err := startGeckoDriverServer(configs.GeckoDriver.BinaryPath, configs.GeckoDriver.Port)
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	err = stopGeckoDriverServer(gds)
	if err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}
