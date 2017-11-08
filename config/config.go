package config

import (
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	BotApi struct {
		ID    int    `required:"true"`
		Key   string `required:"true"`
		Debug bool
	}
	UserApi struct {
		Host      string `required:"true"`
		Port      int    `required:"true"`
		PublicKey string `required:"true"`
		ApiId     int    `required:"true"`
		ApiHash   string `required:"true"`
		Debug     int
	}
	Management struct {
		OwnerUsername   string `required:"true" env:"MANAGEMENT_OWNER_USERNAME"`
		ChannelUsername string `required:"true" env:"MANAGEMENT_CHANNEL_USERNAME"`
	}
	Mongo struct {
		Host string `required:"true"`
		Port int    `required:"true"`
	}
}

func LoadConfig(file string) (*Config, error) {
	var config Config
	err := configor.New(&configor.Config{ENVPrefix: "-"}).Load(&config, file)
	config.UserApi.PublicKey = strings.Replace(config.UserApi.PublicKey, "\\n", "\n", -1)
	return &config, err
}
