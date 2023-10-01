package scraping

import (
	"fmt"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func GetWebDriver(firefoxPath string, driverPort int) (selenium.WebDriver, error) {
	// Connect to the GeckoDriver server running locally
	caps := selenium.Capabilities{"browserName": "firefox"}
	args := []string{"--headless"}
	firefoxCaps := firefox.Capabilities{
		Binary: firefoxPath,
		Args:   args,
	}
	caps.AddFirefox(firefoxCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://dashboard-geckodriver:%d", driverPort))
	if err != nil {
		return nil, err
	}

	return wd, nil
}
