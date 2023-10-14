package scraping

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

var geckoDriverPoolStartPort = 30000
var geckoDriverPool GeckoDriverPool

func NewGeckoDriverPool(geckoDriverPath string, size int) (*GeckoDriverPool, error) {
	if !reflect.DeepEqual(geckoDriverPool, GeckoDriverPool{}) {
		return &geckoDriverPool, nil
	}

	// Create pool
	geckoDriverPool = GeckoDriverPool{
		pool: make(map[int]*GeckoDriverServer, size),
		size: size,
	}
	nextPort := geckoDriverPoolStartPort
	timeout := 20

	for i := 0; i < size; {
		if timeout == 0 {
			closeErrs := stopGeckoDrivers(geckoDriverPool.pool) // Stop open geckodriver instances
			if closeErrs != nil {
				return nil, closeErrs
			}
			return nil, fmt.Errorf("unable to create geckodriver pool. All 20 tested ports are in use")
		}

		// Check port
		available, err := isPortAvailable(nextPort)
		if err != nil {
			closeErrs := stopGeckoDrivers(geckoDriverPool.pool)
			if closeErrs != nil {
				return nil, closeErrs
			}
			return nil, err
		}
		if !available {
			timeout--
			nextPort++
			continue
		}

		// Add to pool
		gds := NewGeckoDriverServer(geckoDriverPath, nextPort)
		err = gds.start()
		if err != nil {
			closeErrs := stopGeckoDrivers(geckoDriverPool.pool)
			if closeErrs != nil {
				return nil, closeErrs
			}
			return nil, err
		}

		geckoDriverPool.pool[nextPort] = gds
		geckoDriverPool.ports = append(geckoDriverPool.ports, nextPort)
		i++
	}

	return &geckoDriverPool, nil
}

type GeckoDriverPool struct {
	// A map of port to geckodriver server
	pool  map[int]*GeckoDriverServer
	ports []int
	size  int
}

func (gdp *GeckoDriverPool) StopAll() error {
	err := stopGeckoDrivers(gdp.pool)
	if err != nil {
		return err
	}

	return nil
}

func stopGeckoDrivers(instances map[int]*GeckoDriverServer) error {
	var errorsStr string
	var errorsN int

	for _, instance := range instances {
		err := instance.Stop()
		if err != nil {
			errorsStr = fmt.Sprintf("%sError %d: %s; ", errorsStr, errorsN, err.Error())
		}
	}

	if errorsStr == "" {
		return nil
	}
	errors := fmt.Errorf("errors occured while stopping the geckodriver instances: %s", errorsStr)

	return errors
}

func (gdp *GeckoDriverPool) WaitGet() (*GeckoDriverServer, error) {
	// Wait for an available GeckoDriver instance
	if len(gdp.pool) == 0 {
		return nil, fmt.Errorf("empty pool")
	}
	for {
		for _, instance := range gdp.pool {
			if instance.busy {
				continue
			} else {
				instance.busy = true
				return instance, nil
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (gdp *GeckoDriverPool) List() []string {
	var instancesAddr []string
	for _, instance := range gdp.pool {
		instancesAddr = append(instancesAddr, instance.addr)
	}

	return instancesAddr
}

func NewGeckoDriverServer(geckoDriverPath string, port int) *GeckoDriverServer {
	localAddr := fmt.Sprintf("http://localhost:%d", port)

	return &GeckoDriverServer{
		GeckoDriverPath: geckoDriverPath,
		Port:            port,
		addr:            localAddr,
	}
}

type GeckoDriverServer struct {
	GeckoDriverPath string
	Port            int
	// Indicates whether the server is being used or not
	busy       bool
	addr       string
	service    *selenium.Service
	mutex      sync.Mutex
	startError error
}

func (gds *GeckoDriverServer) start() error {
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

func (gds *GeckoDriverServer) Release() {
	gds.mutex.Lock()
	defer gds.mutex.Unlock()

	gds.busy = false
}

func isPortAvailable(port int) (bool, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		if strings.HasSuffix(err.Error(), "address already in use") {
			return false, nil
		}
		return false, err
	}
	defer listener.Close()

	return true, nil
}

func GetWebDriver(firefoxPath string) (selenium.WebDriver, *GeckoDriverServer, error) {
	// Get driver
	driver, err := geckoDriverPool.WaitGet()
	if err != nil {
		return nil, driver, err
	}

	// Connect to the GeckoDriver server running locally
	caps := selenium.Capabilities{"browserName": "firefox"}
	args := []string{"--headless"}
	firefoxCaps := firefox.Capabilities{
		Binary: firefoxPath,
		Args:   args,
	}
	caps.AddFirefox(firefoxCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf(driver.addr))
	if err != nil {
		return nil, driver, err
	}

	return wd, driver, nil
}
