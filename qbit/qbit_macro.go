package qbit

import (
	"github.com/sclevine/agouti"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger(utils.GetPackageName())

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type QbitMacro struct {
	driver   *agouti.WebDriver
	page     *agouti.Page
	url      string
	userName string
	password string
}

func DeleteTorrentSeed() error {
	obj := &QbitMacro{}

	if err := obj.setQbitAccount(); err != nil {
		return err
	}

	return obj.deleteTorrentSeed()
}

func (q *QbitMacro) setQbitAccount() error {
	var err error

	q.url, err = utils.GetConfigValue("qbittorrent", "url")
	if err != nil {
		goto ERR
	}

	q.userName, err = utils.GetConfigValue("qbittorrent", "username")
	if err != nil {
		goto ERR
	}

	q.password, err = utils.GetConfigValue("qbittorrent", "password")
	if err != nil {
		goto ERR
	}

	return nil

ERR:
	log.Error(err)
	return err
}

func (q *QbitMacro) deleteTorrentSeed() error {
	var err error

	log.Debug("Call deleteTorrentSeed")
	if err = q.getPage(); err != nil {
		log.Error(err)
		return err
	}
	defer q.putPage()

	if err = q.page.All("#username").At(0).Fill(q.userName); err != nil {
		return err
	}
	if err = q.page.All("#password").At(0).Fill(q.password); err != nil {
		return err
	}

	if err = q.page.All("#login").At(0).Click(); err != nil {
		return err
	}
	q.page.Session().SetImplicitWait(500)

	if err = q.page.All("#completed_filter").At(0).Click(); err != nil {
		return err
	}

	data := q.page.All(".torrentsTableContextMenuTarget")
	completedSeedNum, err := data.Count()
	if err != nil {
		return err
	}

	log.Debug("Number of seeds completed : ", completedSeedNum)
	if completedSeedNum != 0 {
		if err = data.At(0).Click(); err != nil {
			return err
		}
		if err = q.page.All("#deleteButton").At(0).Click(); err != nil {
			return err
		}

		iframe := q.page.All("#confirmDeletionPage_iframe")
		if err = iframe.At(0).SwitchToFrame(); err != nil {
			return err
		}

		if err = q.page.All("#confirmBtn").At(0).Click(); err != nil {
			return err
		}
		log.Debug("Done, Delete completed seed")
	}

	return nil
}

func (q *QbitMacro) getPage() error {
	log.Debug("Call getPage")

	q.driver = agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}),
	)
	if err := q.driver.Start(); err != nil {
		log.Error(err)
		return err
	}

	page, err := q.driver.NewPage()
	if err != nil {
		log.Error(err)
		q.driver.Stop()
		return err
	}

	if err := page.Navigate(q.url); err != nil {
		log.Error(err)
		q.driver.Stop()
		return err
	}

	q.page = page

	return nil
}

func (q *QbitMacro) putPage() {
	log.Debug("Call putPage")

	if err := q.page.CloseWindow(); err != nil {
		log.Error(err)
	}
	if err := q.driver.Stop(); err != nil {
		log.Error(err)
	}
}
