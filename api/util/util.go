package util

import (
	"reflect"

	"github.com/spf13/viper"
)

type Configs struct {
	GeckoDriver GeckoDriverConfigs `mapstructure:"geckodriver"`
	Firefox     FirefoxConfigs     `mapstructure:"firefox"`
	Notion      NotionConfigs      `mapstructure:"notion"`
}

type GeckoDriverConfigs struct {
	BinaryPath string `mapstructure:"binary_path"`
	Port       int    `mapstructure:"port"`
}

type FirefoxConfigs struct {
	BinaryPath string `mapstructure:"binary_path"`
}

type NotionConfigs struct {
	Token         string               `mapstructure:"token"`
	GamesTracker  GamesTrackerConfigs  `mapstructure:"games_tracker"`
	MediasTracker MediasTrackerConfigs `mapstructure:"medias_tracker"`
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
