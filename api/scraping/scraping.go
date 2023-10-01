package scraping

import (
	"fmt"
	"log"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func GetWebDriver(firefoxPath string, driverPort int) (selenium.WebDriver, error) {
	// Connect to the GeckoDriver server running locally
	log.Println(firefoxPath)
	caps := selenium.Capabilities{"browserName": "firefox"}
	args := []string{"--headless"}
	firefoxCaps := firefox.Capabilities{
		Binary: firefoxPath,
		Args:   args,
	}
	caps.AddFirefox(firefoxCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://dashboard-geckodriver:%d", driverPort))
	if err != nil {
		log.Printf("Error in the NewRemote: %s", err.Error())
		return nil, err
	}

	return wd, nil
}
