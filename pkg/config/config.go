package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	PineconeApiKey string `mapstructure:"PINECONEAPIKEY" validate:"required"`
	PineconeIndex  string `mapstructure:"PINECONEINDEX" validate:"required"`
	PineconeHost   string `mapstructure:"PINECONEHOST" validate:"required,url"`
	ApiKey         string `mapstructure:"APIKEY" validate:"required"`
	GeminiApiKey   string `mapstructure:"GEMINIAPIKEY" validate:"required"`
}

var envs = []string{
	"PINECONEAPIKEY", "PINECONEINDEX", "PINECONEHOST", "APIKEY", "GEMINIAPIKEY",
}

func LoadConfig() (Config, error) {
	var config Config

	viper.AutomaticEnv() // Read environment variables

	// Bind environment variables explicitly
	for _, env := range envs {
		if err := viper.BindEnv(env); err != nil {
			return config, err
		}
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	// Validate struct fields
	if err := validator.New().Struct(&config); err != nil {
		return config, err
	}

	return config, nil
}
