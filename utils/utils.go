package utils

import (
	"errors"
	"os"
	"path"
)

func GetConfigDir() (string, error) {
	configDir := path.Join(os.Getenv("HOME"), ".config/go-telegram-bot")

	info, err := os.Stat(configDir)
	if (err != nil) || !info.IsDir() {
		return "", errors.New("It is not a directory")
	}

	return configDir, nil
}
