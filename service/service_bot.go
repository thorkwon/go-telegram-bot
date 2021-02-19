package service

import (
	"encoding/json"
	"fmt"
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

var log = utils.GetLogger(utils.GetPackageName())

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type chatInfo struct {
	ID       int64  `json:"id"`
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
	flagCmd      int
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
	s.chats[chat.ID] = &chatInfo{ID: chat.ID, ChatType: chat.Type, UserName: chat.UserName}

	log.Debug("set chat id:", s.chats[chat.ID])
	log.Debug("set chat :", s.chats)

	// data, _ := json.MarshalIndent(s.chats, "", "\t")
	data, _ := json.Marshal(s.chats)
	log.Debug(string(data))

	if err := ioutil.WriteFile(s.configDir+"/chat_list", data, 0644); err != nil {
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
	s.TouchedFile = true

	if err := ioutil.WriteFile(s.saveFilePath, []byte(msg), 0644); err != nil {
		s.TouchedFile = false
		log.Error(err)
	}
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

	cmd := update.Message.Text
	log.Debugf("cmd msg [%s]", cmd)

	switch cmd {
	case "/cancel":
		s.flagCmd = 0
	case "/debug":
		s.flagCmd = 1
		msg := fmt.Sprintf("[ on/off debug mode ]\nTell to me command:\ne.g.\non main (OR off main)")
		s.SendMsg(update.Message.Chat.ID, msg, true, 60)
		s.SendMsg(update.Message.Chat.ID, utils.GetDebugStatus(), true, 60)
	}

	s.AutoDeleteMsg(update.Message.Chat.ID, update.Message.MessageID, 60)
}

func (s *ServiceBot) downloadFile(fileFullPath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	log.Debug("save file path : ", fileFullPath)
	out, err := os.Create(fileFullPath)
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

	if update.Message.Document != nil {
		log.Debugf("Document %#v", update.Message.Document)

		fileID = update.Message.Document.FileID
		fileName = update.Message.Document.FileName
	} else if update.Message.Photo != nil {
		log.Debugf("Photo %#v", update.Message.Photo)

		fileID = (*update.Message.Photo)[len(*update.Message.Photo)-1].FileID
	} else if update.Message.Video != nil {
		log.Debugf("Video %#v", update.Message.Video)

		fileID = update.Message.Video.FileID
	} else if update.Message.Audio != nil {
		log.Debugf("Audio %#v", update.Message.Audio)

		fileID = update.Message.Audio.FileID
	}

	if fileID != "" {
		fileURL, err := s.bot.GetFileDirectURL(fileID)
		if err != nil {
			log.Error(err)
			goto ERR
		}
		log.Debug("url : ", fileURL)

		var fileFullPath string
		if fileName != "" {
			arr := strings.Split(fileName, ".")
			if arr[len(arr)-1] == "torrent" {
				fileFullPath = s.torrentDir + "/" + fileName
			} else {
				fileFullPath = s.downloadDir + "/" + fileName
			}
		} else if fileName == "" {
			arr := strings.Split(fileURL, "/")
			arr = strings.Split(arr[len(arr)-1], ".")
			times := time.Now().Format("060102_150405")

			fileName = "file_" + times + "." + arr[len(arr)-1]
			fileFullPath = s.downloadDir + "/" + fileName
		}

		if err := s.downloadFile(fileFullPath, fileURL); err != nil {
			log.Error(err)
			goto ERR
		}
	}

	s.SendMsg(update.Message.Chat.ID, "Upload Success", true, 60)
	s.AutoDeleteMsg(update.Message.Chat.ID, update.Message.MessageID, 60)

	return

ERR:
	s.SendMsg(update.Message.Chat.ID, "Upload Failed", true, 60)
}

func (s *ServiceBot) textHandler(update tgbotapi.Update) {
	log.Debug("Call text handler")

	msg := update.Message.Text
	log.Debugf("msg [%s]", msg)

	switch s.flagCmd {
	case 0: // Normal mode
		s.saveMsgToFile(msg)
		s.SendMsg(update.Message.Chat.ID, "Text saved", true, 60)
	case 1: // Debug mode
		arr := strings.Split(msg, " ")
		if len(arr) == 2 {
			if arr[0] == "on" {
				utils.EnableDebugLog(arr[1])
			} else {
				utils.DisableDebugLog(arr[1])
			}
			s.SendMsg(update.Message.Chat.ID, utils.GetDebugStatus(), true, 60)
		}

		s.flagCmd = 0
	}

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

	if err = s.setAdminUser(); err != nil {
		log.Error(err)
		return err
	}

	if err = s.setMsgSaveFile(); err != nil {
		log.Error(err)
		return err
	}

	if err = s.setDownloadDir(); err != nil {
		log.Error(err)
		return err
	}

	if err = s.setTorrentDir(); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *ServiceBot) Stop() {
	s.workQueue.Stop()
	log.Info("Stop service bot")
}
