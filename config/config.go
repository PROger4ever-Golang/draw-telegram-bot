package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Telegram struct {
		BotApi struct {
			ID  int    `json:"id"`
			Key string `json:"key"`
		} `json:"botApi"`
		UserApi struct {
			Host      string `json:"host"`
			Port      int    `json:"port"`
			PublicKey string `json:"publicKey"`
			ApiId     int    `json:"apiId"`
			ApiHash   string `json:"apiHash"`
		} `json:"userApi"`
	} `json:"telegram"`
	Mongo struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"mongo"`
}

func LoadConfig(file string) (Config, error) {
	var config Config

	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return Config{}, err
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
