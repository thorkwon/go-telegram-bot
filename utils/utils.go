package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var loggers map[string]*log.Logger
var logFile *os.File

func init() {
	loggers = make(map[string]*log.Logger)
}

func GetConfigDir() (string, error) {
	configDir := path.Join(os.Getenv("HOME"), ".config/go-telegram-bot")

	info, err := os.Stat(configDir)
	if (err != nil) || !info.IsDir() {
		return "", errors.New("It is not a directory")
	}

	return configDir, nil
}

func GetConfigData(configFile string) (string, error) {
	configDir, err := GetConfigDir()
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

func getFileAndLine(path string, line int) string {
	arr := strings.Split(path, "/")
	filePath := fmt.Sprintf("%s:%d", arr[len(arr)-1], line)

	return filePath
}

func GetLogger(key string) *log.Logger {
	logger, isKey := loggers[key]
	if isKey {
		return logger
	}

	if logFile == nil {
		configDir, err := GetConfigDir()
		if err != nil {
			return nil
		}

		fpLog, err := os.OpenFile(configDir+"/bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil
		}

		logFile = fpLog
	}

	loggers[key] = log.New()

	loggers[key].SetReportCaller(true)

	formatter := &log.TextFormatter{
		ForceColors: true,
		// DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%-30s", getFileAndLine(f.File, f.Line))
		},
	}
	loggers[key].SetFormatter(formatter)

	loggers[key].SetOutput(logFile)
	// loggers[key].SetOutput(os.Stdout)
	loggers[key].SetLevel(log.InfoLevel)

	return loggers[key]
}

func EnableDebugLog(key string) {
	logger, isKey := loggers[key]
	if isKey {
		logger.SetLevel(log.DebugLevel)
	}
}
