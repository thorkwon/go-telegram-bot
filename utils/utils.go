package utils

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
)

func getConfigDir() (string, error) {
	configDir := path.Join(os.Getenv("HOME"), ".config/go-telegram-bot")

	info, err := os.Stat(configDir)
	if (err != nil) || !info.IsDir() {
		return "", errors.New("It is not a directory")
	}

	return configDir, nil
}

func GetConfigData(configFile string) (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return configDir, err
	}

	configPath, err := ioutil.ReadFile(configDir + "/" + configFile)
	if err != nil {
		return "", err
	}

	filePath := string(configPath)[:len(configPath)-1]

	return filePath, nil
}
