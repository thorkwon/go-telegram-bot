package service

import (
	"encoding/json"
	"io/ioutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger("service")

func init() {
	// utils.EnableDebugLog("service")
}

type chatInfo struct {
	Id       int64  `json:"id"`
	ChatType string `json:"type"`
	UserName string `json:"name"`
}

type ServiceBot struct {
	bot       *tgbotapi.BotAPI
	updates   tgbotapi.UpdatesChannel
	chats     map[int64]*chatInfo
	adminUser string
}

func NewServiceBot() *ServiceBot {
	obj := &ServiceBot{}
	obj.chats = make(map[int64]*chatInfo)

	return obj
}

func (s *ServiceBot) getToken() (string, error) {
	tokenKey, err := utils.GetConfigData("token_key")
	if err != nil {
		log.Error("ERR :", err)
		return "", err
	}

	return tokenKey, nil
}

func (s *ServiceBot) setChat(chat *tgbotapi.Chat) {
	s.chats[chat.ID] = &chatInfo{Id: chat.ID, ChatType: chat.Type, UserName: chat.UserName}

	log.Debug("set chat id:", s.chats[chat.ID])
	log.Debug("set chat :", s.chats)

	// data, _ := json.MarshalIndent(s.chats, "", "\t")
	data, _ := json.Marshal(s.chats)
	log.Debug(string(data))

	configDir, err := utils.GetConfigDir()
	if err != nil {
		log.Error(err)
	} else {
		err = ioutil.WriteFile(configDir+"/chat_list", data, 0644)
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *ServiceBot) GetChat() map[int64]*chatInfo {
	if len(s.chats) != 0 {
		return s.chats
	}

	configDir, err := utils.GetConfigDir()
	if err != nil {
		log.Error(err)
		return nil
	}

	data, err := ioutil.ReadFile(configDir + "/chat_list")
	if err != nil {
		log.Error(err)
		return nil
	}

	json.Unmarshal(data, &s.chats)
	log.Debug("get chat:", s.chats)

	return s.chats
}

func (s *ServiceBot) setAdminUser() error {
	adminUser, err := utils.GetConfigData("admin_user")
	if err != nil {
		log.Panicln("ERR :", err)
		return err
	}
	s.adminUser = adminUser
	log.Infof("Admin user : [%s]", s.adminUser)

	return nil
}

func (s *ServiceBot) GetAdminUser() string {
	return s.adminUser
}

func (s *ServiceBot) updateReceiver() {
	for update := range s.updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		chats := s.GetChat()
		if chats[update.Message.Chat.ID] == nil {
			s.setChat(update.Message.Chat)
			log.Infof("New chat id, save chat id [%d]", update.Message.Chat.ID)
		}

		log.Debugf("[%s] %s]", update.Message.From.UserName, update.Message.Text)
	}
}

func (s *ServiceBot) SendMsg(chatID int64, msg string, delete bool, delay int) {
	if delete {
		// auto delete msg
	}
	s.bot.Send(tgbotapi.NewMessage(chatID, msg))
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

	s.setAdminUser()

	return nil
}
