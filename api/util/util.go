package util

import (
	"reflect"
	"time"

	"github.com/diogovalentte/notionapi"
	"github.com/go-playground/validator/v10"
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
	Token       string             `mapstructure:"token"`
	GameTracker GameTrackerConfigs `mapstructure:"game_tracker"`
}

type GameTrackerConfigs struct {
	GamesDBID string `mapstructure:"games_db_id"`
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

func IsValidDate(fl validator.FieldLevel) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, fl.Field().String())

	return err == nil
}

func GetNotionClient() (*notionapi.Client, error) {
	configs, err := GetConfigs()
	if err != nil {
		return nil, err
	}

	return notionapi.NewClient(notionapi.Token(configs.Notion.Token)), nil
}
