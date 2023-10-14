package util

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/spf13/viper"
)

type Configs struct {
	Database    DatabaseConfigs    `mapstructure:"database"`
	GeckoDriver GeckoDriverConfigs `mapstructure:"geckodriver"`
	Firefox     FirefoxConfigs     `mapstructure:"firefox"`
}

type DatabaseConfigs struct {
	FolderPath string `mapstructure:"databases_folder_abs_path"`
}

type GeckoDriverConfigs struct {
	BinaryPath string `mapstructure:"binary_path"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type FirefoxConfigs struct {
	BinaryPath string `mapstructure:"binary_path"`
}

type GamesTrackerConfigs struct {
	DBID string `mapstructure:"db_id"`
}

type MediasTrackerConfigs struct {
	DBID string `mapstructure:"db_id"`
}

var configs Configs

func GetConfigsWithoutDefaults(configPath string) (*Configs, error) {
	if !reflect.DeepEqual(configs, Configs{}) {
		return &configs, nil
	}

	viper.SetConfigName("configs")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		return nil, err
	}

	return &configs, nil
}

func GetConfigs() (*Configs, error) {
	defaultConfigPath := "configs"
	return GetConfigsWithoutDefaults(defaultConfigPath)
}

func GetImageFromURL(imageURL string) ([]byte, error) {
	response, err := http.Get(imageURL)
	if err != nil {
		err = fmt.Errorf("error downloading game cover image: %s", err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to download game cover image. Status code: %d", response.StatusCode)
		return nil, err
	}

	imageBytes, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("error reading game cover image data: %s", err)
		return nil, err
	}

	return imageBytes, nil
}
