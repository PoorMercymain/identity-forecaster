package config

import (
	"errors"
	"identity-forecaster/internal/pkg/logger"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

const (
	defaultHost                      = "localhost"
	defaultContainerHost             = "0.0.0.0"
	defaultPort                      = "8787"
	defaultDSN                       = "host=localhost dbname=identity-forecaster user=identity-forecaster password=identity-forecaster port=5432 sslmode=disable"
	defaultContainerDSN              = "host=postgres dbname=identity-forecaster user=identity-forecaster password=identity-forecaster port=5432 sslmode=disable"
	defaultLogfile                   = "logfile.log"
	defaultAPIs                      = "https://api.agify.io/,https://api.genderize.io/,https://api.nationalize.io/"
	defaultRetriesAmount             = 5
	defaultRetryIntervalMilliseconds = 150
	defaultIsInContainer             = false
	envFile                          = ".env"
)

type Config struct {
	ServiceHost               string `env:"HOST"`
	ServicePort               string `env:"PORT"`
	DatabaseDSN               string `env:"DSN"`
	Logfile                   string `env:"LOGFILE"`
	APIsStr                   string `env:"API"`
	RetriesAmount             uint   `env:"RETRIES"`
	RetryIntervalMilliseconds uint   `env:"INTERVAL"`
	IsInContainer             bool   `env:"IN_CONTAINER"`
	APIs                      []string
}

func LoadConfig() *Config {
	err := godotenv.Load(envFile)

	var envCfg = Config{}
	var alreadyInitialized bool
	if val, wasIsInContainerSet := os.LookupEnv("IN_CONTAINER"); wasIsInContainerSet {
		alreadyInitialized = wasIsInContainerSet
		if val == "true" {
			envCfg = Config{
				ServiceHost:               defaultContainerHost,
				ServicePort:               defaultPort,
				DatabaseDSN:               defaultContainerDSN,
				Logfile:                   defaultLogfile,
				RetriesAmount:             defaultRetriesAmount,
				RetryIntervalMilliseconds: defaultRetryIntervalMilliseconds,
				IsInContainer:             defaultIsInContainer,
				APIsStr:                   defaultAPIs,
			}
		} else {
			envCfg = Config{
				ServiceHost:               defaultHost,
				ServicePort:               defaultPort,
				DatabaseDSN:               defaultDSN,
				Logfile:                   defaultLogfile,
				RetriesAmount:             defaultRetriesAmount,
				RetryIntervalMilliseconds: defaultRetryIntervalMilliseconds,
				IsInContainer:             defaultIsInContainer,
				APIsStr:                   defaultAPIs,
			}
		}
	}

	if !alreadyInitialized {
		envCfg = Config{
			ServiceHost:               defaultHost,
			ServicePort:               defaultPort,
			DatabaseDSN:               defaultDSN,
			Logfile:                   defaultLogfile,
			RetriesAmount:             defaultRetriesAmount,
			RetryIntervalMilliseconds: defaultRetryIntervalMilliseconds,
			IsInContainer:             defaultIsInContainer,
			APIsStr:                   defaultAPIs,
		}
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	err = env.Parse(&envCfg)
	if err != nil {
		panic(err)
	}

	envCfg.APIs = strings.Split(envCfg.APIsStr, ",")

	logger.Logger().Infoln(envCfg)
	return &envCfg
}
