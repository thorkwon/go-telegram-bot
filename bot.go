package main // import "github.com/thorkwon/go-telegram-bot"

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/thorkwon/go-telegram-bot/service"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger("main")

func init() {
	// utils.EnableDebugLog("main")
}

func main() {
	service := service.NewServiceBot()

	err := service.Start()
	if err != nil {
		log.Panicln(err)
	}

	// Exit
	doneExit := make(chan os.Signal)
	signal.Notify(doneExit, os.Interrupt, syscall.SIGTERM)

	<-doneExit
	//log.Println("stop service bot")
	log.Info("Stop service bot")
}
