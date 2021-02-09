package service

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/thorkwon/go-telegram-bot/service/queue"
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
	bot          *tgbotapi.BotAPI
	updates      tgbotapi.UpdatesChannel
	configDir    string
	chats        map[int64]*chatInfo
	adminUser    string
	saveFilePath string
	downloadDir  string
	torrentDir   string
	TouchedFile  bool
	workQueue    *queue.WorkQueue
}

func NewServiceBot() *ServiceBot {
	obj := &ServiceBot{}
	obj.chats = make(map[int64]*chatInfo)
	obj.workQueue = queue.NewWorkQueue(obj.deleteMsg)

	return obj
}

func (s *ServiceBot) getToken() (string, error) {
	tokenKey, err := utils.GetConfigData("token_key")
	if err != nil {
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

	err := ioutil.WriteFile(s.configDir+"/chat_list", data, 0644)
	if err != nil {
		log.Error(err)
	}
}

func (s *ServiceBot) GetChat() map[int64]*chatInfo {
	if len(s.chats) != 0 {
		return s.chats
	}

	data, err := ioutil.ReadFile(s.configDir + "/chat_list")
	if err != nil {
		log.Error(err)
	}

	json.Unmarshal(data, &s.chats)
	log.Debug("get chat:", s.chats)

	return s.chats
}

func (s *ServiceBot) setAdminUser() error {
	adminUser, err := utils.GetConfigData("admin_user")
	if err != nil {
		return err
	}
	s.adminUser = adminUser
	log.Infof("Admin user : [%s]", s.adminUser)

	return nil
}

func (s *ServiceBot) GetAdminUser() string {
	return s.adminUser
}

func (s *ServiceBot) setDownloadDir() error {
	dir, err := utils.GetConfigData("download_dir")
	if err != nil {
		return err
	}

	s.downloadDir = dir
	log.Info("Download dir path : ", s.downloadDir)

	return err
}

func (s *ServiceBot) setTorrentDir() error {
	dir, err := utils.GetConfigData("torrent_dir")
	if err != nil {
		return err
	}

	s.torrentDir = dir
	log.Info("Torrent seed dir path : ", s.torrentDir)

	return err
}

func (s *ServiceBot) setMsgSaveFile() error {
	watchFile, err := utils.GetConfigData("watch_file")
	if err != nil {
		return err
	}

	s.saveFilePath = watchFile
	log.Info("Save file path : ", s.saveFilePath)

	return err
}

func (s *ServiceBot) saveMsgToFile(msg string) {
	err := ioutil.WriteFile(s.saveFilePath, []byte(msg), 0644)
	if err != nil {
		log.Error(err)
	}

	s.TouchedFile = true
}

func (s *ServiceBot) deleteMsg(chatID int64, msgID int) {
	s.bot.DeleteMessage(tgbotapi.NewDeleteMessage(chatID, msgID))
}

func (s *ServiceBot) AutoDeleteMsg(chatID int64, msgID int, delay int) {
	s.workQueue.AddTask(chatID, msgID, delay)
}

func (s *ServiceBot) SendMsg(chatID int64, msg string, delete bool, delay int) {
	ret, err := s.bot.Send(tgbotapi.NewMessage(chatID, msg))
	if err == nil && delete {
		s.AutoDeleteMsg(chatID, ret.MessageID, delay)
	}
}

func (s *ServiceBot) cmdHandler(update tgbotapi.Update) {
	log.Debug("Call cmd handler")
}

func (s *ServiceBot) downloadFile(fileName string, url string) error {
	if fileName == "" {
		arr := strings.Split(url, "/")
		arr = strings.Split(arr[len(arr)-1], ".")
		times := time.Now().Format("060102_150405")

		fileName = "file_" + times + "." + arr[len(arr)-1]
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	log.Debug("save file path : ", s.downloadDir+"/"+fileName)
	out, err := os.Create(s.downloadDir + "/" + fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (s *ServiceBot) fileHandler(update tgbotapi.Update) {
	log.Debug("Call file handler")

	var fileID string
	var fileName string

	if update.Message.Photo != nil {
		log.Debugf("Photo %#v", update.Message.Photo)

		fileID = (*update.Message.Photo)[len(*update.Message.Photo)-1].FileID
	}
	if update.Message.Video != nil {
		log.Debugf("Video %#v", update.Message.Video)

		fileID = update.Message.Video.FileID
	}
	if update.Message.Audio != nil {
		log.Debugf("Audio %#v", update.Message.Audio)

		fileID = update.Message.Audio.FileID
	}
	if update.Message.Document != nil {
		log.Debugf("Document %#v", update.Message.Document)

		fileID = update.Message.Document.FileID
		fileName = update.Message.Document.FileName
	}

	if fileID != "" {
		fileURL, err := s.bot.GetFileDirectURL(fileID)
		if err != nil {
			log.Error(err)
			goto ERR
		}
		log.Debug("url : ", fileURL)

		err = s.downloadFile(fileName, fileURL)
		if err != nil {
			log.Error(err)
			goto ERR
		}
	}

	s.SendMsg(update.Message.Chat.ID, "Upload Success", true, 60)

	return

ERR:
	s.SendMsg(update.Message.Chat.ID, "Upload Failed", true, 60)
}

func (s *ServiceBot) textHandler(update tgbotapi.Update) {
	log.Debug("Call text handler")

	log.Debugf("msg [%s]", update.Message.Text)
	s.saveMsgToFile(update.Message.Text)
	s.SendMsg(update.Message.Chat.ID, "Text saved", true, 60)
	s.AutoDeleteMsg(update.Message.Chat.ID, update.Message.MessageID, 60)
}

func (s *ServiceBot) updateReceiver() {
	for update := range s.updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if s.adminUser != update.Message.From.UserName {
			log.Warn("Not admin user : ", update.Message.From.UserName)
			continue
		}

		chats := s.GetChat()
		if chats[update.Message.Chat.ID] == nil {
			s.setChat(update.Message.Chat)
			log.Infof("New chat id, save chat id [%d]", update.Message.Chat.ID)
		}

		if update.Message.IsCommand() {
			go s.cmdHandler(update)
		} else if update.Message.Photo == nil && update.Message.Video == nil && update.Message.Audio == nil && update.Message.Document == nil {
			go s.textHandler(update)
		} else {
			go s.fileHandler(update)
		}
	}
}

func (s *ServiceBot) Start() error {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		log.Error(err)
		return err
	}
	s.configDir = configDir

	tokenKey, err := s.getToken()
	if err != nil {
		log.Error(err)
		return err
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
	if err != nil {
		log.Error(err)
		return err
	}
	s.updates = updates

	go s.updateReceiver()
	log.Info("Start service bot")

	err = s.setAdminUser()
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.setMsgSaveFile()
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.setDownloadDir()
	if err != nil {
		log.Error(err)
		return err
	}

	err = s.setTorrentDir()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *ServiceBot) Stop() {
	s.workQueue.Stop()
	log.Info("Stop service bot")
}
