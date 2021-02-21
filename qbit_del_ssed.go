package main // import "github.com/thorkwon/go-telegram-bot"

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/thorkwon/go-telegram-bot/qbit"
	"github.com/thorkwon/go-telegram-bot/utils"
	"github.com/thorkwon/go-telegram-bot/watch"
)

var log = utils.GetLogger(utils.GetPackageName())

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

func deleteTorrentSeed(seedName string, arg interface{}) {
	log.Debug("call deleteTorrentSeed")
	log.Debug("seedname :", seedName)

	if err := qbit.DeleteTorrentSeed(); err != nil {
		log.Error(err)
	}
}

func main() {
	downloadWatcher := watch.DownloadPolling(deleteTorrentSeed, nil)

	// Exit
	doneExit := make(chan os.Signal)
	signal.Notify(doneExit, os.Interrupt, syscall.SIGTERM)

	<-doneExit

	if downloadWatcher != nil {
		downloadWatcher.StopPolling()
	}
}
