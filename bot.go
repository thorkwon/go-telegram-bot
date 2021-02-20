package main // import "github.com/thorkwon/go-telegram-bot"

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/thorkwon/go-telegram-bot/crawling"
	"github.com/thorkwon/go-telegram-bot/qbit"
	"github.com/thorkwon/go-telegram-bot/service"
	"github.com/thorkwon/go-telegram-bot/utils"
	"github.com/thorkwon/go-telegram-bot/watch"
)

var log = utils.GetLogger(utils.GetPackageName())

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type infoArg struct {
	service *service.ServiceBot
	chatID  int64
}

func sendClipboardToChat(data string, arg interface{}) {
	var info *infoArg = arg.(*infoArg)

	log.Debug("call sendClipboardToChat")

	if info.service.TouchedFile {
		info.service.TouchedFile = false
		return
	}

	log.Debug("data :", data)
	info.service.SendMsg(info.chatID, data, false, 0)
}

func deleteTorrentSeed(seedName string, arg interface{}) {
	var info *infoArg = arg.(*infoArg)

	log.Debug("call deleteTorrentSeed")
	log.Debug("seedname :", seedName)

	msg := fmt.Sprintf("Torrent download complete!!!\n[%s]", seedName)
	info.service.SendMsg(info.chatID, msg, true, 3600)

	if err := qbit.DeleteTorrentSeed(); err != nil {
		log.Error(err)
	}
}

func sendCOVID19Status(data string, arg interface{}) {
	var info *infoArg = arg.(*infoArg)

	log.Debug("call sendCOVID19Status")

	info.service.SendMsg(info.chatID, data, false, 0)
}

func sendCoinPremium(data string, arg interface{}) {
	var info *infoArg = arg.(*infoArg)

	log.Debug("call sendCoinPremium")

	info.service.SendMsg(info.chatID, data, false, 0)
}

func getPrivateChatID(service *service.ServiceBot) int64 {
	chats := service.GetChat()
	adminUser := service.GetAdminUser()
	var ret int64

	for key, val := range chats {
		log.Debug("call getPrivateChatID :", key, val)
		if val.UserName == adminUser && val.ChatType == "private" {
			ret = key
		}
	}

	return ret
}

func main() {
	service := service.NewServiceBot()
	var clipboardWatcher *watch.ClipboardWatcher
	var downloadWatcher *watch.DownloadWatcher
	var covidCrawler *crawling.COVID19Crawler
	var coinCrawler *crawling.CoinCrawler

	if err := service.Start(); err != nil {
		log.Fatal(err)
	}

	chatID := getPrivateChatID(service)
	if chatID == 0 {
		log.Warn("No such private chat")
		log.Warn("Clipboard watching service has not started")
		log.Warn("Download watching service has not started")
		log.Warn("COVID-19 notice service has not started")
		log.Warn("Coin premium notice service has not started")
	} else {
		info := &infoArg{service: service, chatID: chatID}
		clipboardWatcher = watch.ClipboardPolling(sendClipboardToChat, info)
		downloadWatcher = watch.DownloadPolling(deleteTorrentSeed, info)
		covidCrawler = crawling.NoticeCOVID19Status(sendCOVID19Status, info)
		coinCrawler = crawling.NoticeCoinPremium(sendCoinPremium, info)
	}

	// Exit
	doneExit := make(chan os.Signal)
	signal.Notify(doneExit, os.Interrupt, syscall.SIGTERM)

	<-doneExit

	if clipboardWatcher != nil {
		clipboardWatcher.StopPolling()
	}
	if downloadWatcher != nil {
		downloadWatcher.StopPolling()
	}
	if covidCrawler != nil {
		covidCrawler.Stop()
	}
	if coinCrawler != nil {
		coinCrawler.Stop()
	}

	service.Stop()
}
