package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/bigkevmcd/go-configparser"
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

func GetConfigValue(section string, option string) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return configDir, err
	}

	configFile := path.Join(configDir, "go-telegram-bot.conf")

	configParser, err := configparser.NewConfigParserFromFile(configFile)
	if err != nil {
		return "", err
	}

	isValue, _ := configParser.HasOption(section, option)
	if !isValue {
		return "", errors.New("No such section or option")
	}

	value, _ := configParser.Get(section, option)

	return value, nil
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

func DisableDebugLog(key string) {
	logger, isKey := loggers[key]
	if isKey {
		logger.SetLevel(log.InfoLevel)
	}
}

func GetDebugStatus() string {
	var status string

	for key := range loggers {
		logger, _ := loggers[key]
		status = status + fmt.Sprintf("%-8s%s\n", logger.GetLevel().String(), key)
	}

	return status
}

func GetPackageName() string {
	pc, _, _, _ := runtime.Caller(1)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	pkage := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		pkage = strings.Join(parts[0:pl-2], ".")
	} else {
		pkage = strings.Join(parts[0:pl-1], ".")
	}

	arr := strings.Split(pkage, "/")
	pkage = strings.Split(arr[len(arr)-1], ".")[0]

	return pkage
}
