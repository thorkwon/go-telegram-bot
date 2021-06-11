package watch

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/thorkwon/go-telegram-bot/utils"
)

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type DownloadWatcher struct {
	donePolling   chan bool
	isDonePolling chan bool
	watcher       struct{}
	cb            func(string, interface{})
	arg           interface{}
}

func DownloadPolling(cb func(string, interface{}), arg interface{}) *DownloadWatcher {
	obj := &DownloadWatcher{donePolling: make(chan bool), isDonePolling: make(chan bool)}

	filePath, err := utils.GetConfigData("qbit_downloads_dir")
	if err != nil {
		log.Error(err)
		return nil
	}

	obj.setCbFunc(cb, arg)

	go obj.pollingProcess(filePath)

	return obj
}

func (c *DownloadWatcher) StopPolling() {
	c.donePolling <- true

	<-c.isDonePolling
}

func (c *DownloadWatcher) setCbFunc(cb func(string, interface{}), arg interface{}) {
	c.cb = cb
	c.arg = arg
}

func (c *DownloadWatcher) pollingProcess(pollingPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panic(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Debug("Watcher event : ", event)
					arrName := strings.Split(event.Name, "/")
					if !strings.HasPrefix(arrName[len(arrName)-1], ".") {
						c.cb(arrName[len(arrName)-1], c.arg)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Error(err)
					return
				}
			}
		}
	}()

	if err = watcher.Add(pollingPath); err != nil {
		log.Panic(err)
	}
	log.Info("Download watching path : ", pollingPath)

	<-c.donePolling

	log.Info("Stop download watching service")
	c.isDonePolling <- true
}
