package scraping

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

type GeckoDriverServer struct {
	GeckoDriverPath string
	Port            int
	addr            string
	service         *selenium.Service
	mutex           sync.Mutex
	startError      error
}

func NewGeckoDriverServer(geckoDriverPath string, port int) *GeckoDriverServer {
	localAddr := fmt.Sprintf("http://localhost:%d", port)

	return &GeckoDriverServer{
		GeckoDriverPath: geckoDriverPath,
		Port:            port,
		addr:            localAddr,
	}
}

func (gds *GeckoDriverServer) Start() error {
	gds.mutex.Lock()
	defer gds.mutex.Unlock()

	// Start a GeckoDriver WebDriver server instance (if one is not already running)
	log.Println("Starting the GeckoDriver server")
	opts := []selenium.ServiceOption{
		selenium.Output(os.Stderr),
	}
	service, err := selenium.NewGeckoDriverService(gds.GeckoDriverPath, gds.Port, opts...)
	if err != nil {
		gds.startError = err
		return err
	}
	log.Println("GeckoDriver server started")

	gds.service = service

	return nil
}

func (gds *GeckoDriverServer) Wait() error {
	// Wait for the GeckoDriver server status to be OK
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		resp, err := http.Get(gds.addr + "/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		} else {
			return fmt.Errorf("%s; %s", err, gds.startError)
		}
	}

	return fmt.Errorf("server did not respond on port %d", gds.Port)
}

func (gds *GeckoDriverServer) Stop() error {
	gds.mutex.Lock()
	defer gds.mutex.Unlock()
	log.Println("Stopping the GeckoDriver server")

	if gds.service == nil {
		return fmt.Errorf("GeckoDriver server is not running")
	}

	if err := gds.service.Stop(); err != nil {
		return err
	}

	fmt.Println("GeckoDriver server stopped")

	gds.service = nil

	return nil
}

func GetWebDriver(firefoxPath string, driverPort int) (selenium.WebDriver, error) {
	// Connect to the GeckoDriver server running locally
	caps := selenium.Capabilities{"browserName": "firefox"}
	args := []string{"--headless"}
	firefoxCaps := firefox.Capabilities{
		Binary: firefoxPath,
		Args:   args,
	}
	caps.AddFirefox(firefoxCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", driverPort))
	if err != nil {
		return nil, err
	}

	return wd, nil
}
