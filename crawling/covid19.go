package crawling

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/thorkwon/go-telegram-bot/utils"
)

var log = utils.GetLogger("crawling")

func init() {
	// utils.EnableDebugLog("crawling")
}

type COVID19Crawler struct {
	doc    *goquery.Document
	cb     func(string, interface{})
	arg    interface{}
	done   bool
	isDone chan bool
}

func NoticeCOVID19Status(cb func(string, interface{}), arg interface{}) *COVID19Crawler {
	obj := &COVID19Crawler{isDone: make(chan bool)}
	obj.setCbFunc(cb, arg)

	go obj.crawlingProcess()

	return obj
}

func (c *COVID19Crawler) setCbFunc(cb func(string, interface{}), arg interface{}) {
	c.cb = cb
	c.arg = arg
}

func (c *COVID19Crawler) crawlingProcess() {
	flagCheckTime := false
	flagChecked := false
	oldMsg := ""

	log.Info("COVID-19 status notification service start")
	for !c.done {
		now := time.Now()
		if now.Hour() >= 9 && now.Hour() < 18 {
			flagCheckTime = true
		} else {
			flagCheckTime = false
			flagChecked = false
			oldMsg = ""

			continue
		}

		nowMon, _ := strconv.Atoi(time.Now().Format("01"))
		nowDay, _ := strconv.Atoi(time.Now().Format("02"))

		if flagCheckTime && !flagChecked && now.Second() == 0 {
			mon, day := c.checkUpdateCOVID19()
			if nowDay == day && nowMon == mon {
				msg := c.getCOVID19Status()
				if msg != "" {
					flagChecked = true
					c.cb(msg, c.arg)
					oldMsg = msg
				}
			}
		}

		if flagCheckTime && flagChecked && now.Minute() == 0 && now.Second() == 0 {
			log.Debug("Double check COVID-19 status")
			msg := c.getCOVID19Status()
			if oldMsg != msg {
				log.Debug("Renew COVID-19 status")
				c.cb(msg, c.arg)
				oldMsg = msg
			}
		}

		time.Sleep(time.Second)
	}
	c.isDone <- true
}

func (c *COVID19Crawler) Stop() {
	c.done = true
	<-c.isDone
	log.Info("Stop COVID-19 status crawling service")
}

func (c *COVID19Crawler) getURL() error {
	resp, err := http.Get("http://ncov.mohw.go.kr")
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
		log.Error(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	c.doc = doc

	return nil
}

func (c *COVID19Crawler) getCOVID19Status() string {
	if err := c.getURL(); err != nil {
		return ""
	}

	date := c.doc.Find("span.livedate").First().Text()
	arr := strings.Split(date, "(")
	arr = strings.Split(arr[1], " ")
	arr = strings.Split(arr[0], ".")

	mon, _ := strconv.Atoi(arr[0])
	day, _ := strconv.Atoi(arr[1])

	total := c.doc.Find("span.num").First().Text()
	add := c.doc.Find("span.before").First().Text()

	msg := fmt.Sprintf("Update COVID-19 status\n확진환자 (%d/%d)\n%s\n%s", mon, day, total, add)

	return msg
}

func (c *COVID19Crawler) checkUpdateCOVID19() (int, int) {
	if err := c.getURL(); err != nil {
		return 0, 0
	}

	date := c.doc.Find("span.livedate").First().Text()
	arr := strings.Split(date, "(")
	arr = strings.Split(arr[1], " ")
	arr = strings.Split(arr[0], ".")

	mon, _ := strconv.Atoi(arr[0])
	day, _ := strconv.Atoi(arr[1])

	return mon, day
}
