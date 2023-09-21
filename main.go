package main

import (
	"github.com/diogovalentte/dashboard/api"
	"github.com/diogovalentte/dashboard/api/scraping"
	"github.com/diogovalentte/dashboard/api/util"
)

func init() {
	configs, err := util.GetConfigs()
	if err != nil {
		panic(err)
	}

	// Start the GeckoDriver server
	gds := scraping.NewGeckoDriverServer(configs.GeckoDriver.BinaryPath, configs.GeckoDriver.Port)

	go gds.Start()

	err = gds.Wait()
	if err != nil {
		panic(err)
	}
}

func main() {
	router := api.SetupRouter()

	router.Run()
}
