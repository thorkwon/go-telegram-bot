package watch

import (
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger("watch")

func init() {
	// utils.EnableDebugLog("watch")
}

type ClipboardWatcher struct {
	donePolling   chan bool
	isDonePolling chan bool
	watcher       struct{}
	cb            func(string, interface{})
	arg           interface{}
}

func ClipboardPolling(cb func(string, interface{}), arg interface{}) *ClipboardWatcher {
	obj := &ClipboardWatcher{donePolling: make(chan bool), isDonePolling: make(chan bool)}

	filePath, err := utils.GetConfigData("watch_file")
	if err != nil {
		log.Error(err)
		return nil
	}

	obj.setCbFunc(cb, arg)

	go obj.pollingProcess(filePath)

	return obj
}

func (c *ClipboardWatcher) StopPolling() {
	c.donePolling <- true

	<-c.isDonePolling
}

func (c *ClipboardWatcher) setCbFunc(cb func(string, interface{}), arg interface{}) {
	c.cb = cb
	c.arg = arg
}

func (c *ClipboardWatcher) pollingProcess(pollingPath string) {
	var oldData string

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

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Debug("Watcher event : ", event)
					data, err := ioutil.ReadFile(event.Name)
					if err == nil && len(data) != 0 && oldData != string(data) {
						c.cb(string(data)[:len(data)-1], c.arg)
					}
					oldData = string(data)
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
	log.Info("Clipboard watching path : ", pollingPath)

	<-c.donePolling

	log.Info("Stop clipboard watching service")
	c.isDonePolling <- true
}
