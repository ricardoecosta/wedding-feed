package config

import (
	"os"
	"log"
	"encoding/json"
)

type Config struct {
	Port uint32 `json:"port"`
}

func Load(configFile string) Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("couldn't file config file: %v", configFile)
	}

	var config Config
	json.NewDecoder(file).Decode(&config)
	return config
}
