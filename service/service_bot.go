package service

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger("service")

func init() {
	// utils.EnableDebugLog("service")
}

type ServiceBot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func NewServiceBot() *ServiceBot {
	obj := &ServiceBot{}

	return obj
}

func (s *ServiceBot) getToken() (string, error) {
	log.Debug("call get Token")

	tokenKey, err := utils.GetConfigData("token_key")
	if err != nil {
		log.Error("ERR :", err)
		return "", err
	}

	return tokenKey, nil
}

func (s *ServiceBot) updateReceiver() {
	for update := range s.updates {
		log.Debug("==== update:", update)
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Debugf("[%s] %s]", update.Message.From.UserName, update.Message.Text)
	}
}

func (s *ServiceBot) Start() error {
	tokenKey, err := s.getToken()
	if err != nil {
		log.Error("get token err")
		return nil
	}
	log.Debug("Token Key :", tokenKey)

	bot, err := tgbotapi.NewBotAPI(tokenKey)
	if err != nil {
		log.Error(err)
		return err
	}
	// bot.Debug = true
	s.bot = bot

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	s.updates = updates

	go s.updateReceiver()
	log.Info("Start service bot")

	return nil
}

func (s *ServiceBot) setChat() {

}

func (s *ServiceBot) getChat() {

}

func (s *ServiceBot) sendMsg(chatID int64, msg string, delete bool, delay int) {
	if delete {
		// auto delete msg
	}
	s.bot.Send(tgbotapi.NewMessage(chatID, msg))
}
