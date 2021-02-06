package main // import "github.com/thorkwon/go-telegram-bot"

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/thorkwon/go-telegram-bot/service"
	"github.com/thorkwon/go-telegram-bot/utils"
	"github.com/thorkwon/go-telegram-bot/watch"
)

var log = utils.GetLogger("main")

func init() {
	// utils.EnableDebugLog("main")
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

	err := service.Start()
	if err != nil {
		log.Fatal(err)
	}

	chatID := getPrivateChatID(service)
	if chatID == 0 {
		log.Fatal("No such private chat")
	}

	info := &infoArg{service: service, chatID: chatID}
	clipboardWatcher = watch.ClipboardPolling(sendClipboardToChat, info)

	// Exit
	doneExit := make(chan os.Signal)
	signal.Notify(doneExit, os.Interrupt, syscall.SIGTERM)

	<-doneExit

	if clipboardWatcher != nil {
		clipboardWatcher.StopPolling()
	}

	log.Info("Stop service bot")
}
