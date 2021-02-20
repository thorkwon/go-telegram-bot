package crawling

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sclevine/agouti"
	// "github.com/thorkwon/go-telegram-bot/utils"
)

func init() {
	// utils.EnableDebugLog(utils.GetPackageName())
}

type CoinCrawler struct {
	driver *agouti.WebDriver
	page   *agouti.Page
	cb     func(string, interface{})
	arg    interface{}
	done   bool
	isDone chan bool
}

func NoticeCoinPremium(cb func(string, interface{}), arg interface{}) *CoinCrawler {
	obj := &CoinCrawler{isDone: make(chan bool)}
	obj.setCbFunc(cb, arg)

	go obj.crawlingProcess()

	return obj
}

func (c *CoinCrawler) setCbFunc(cb func(string, interface{}), arg interface{}) {
	c.cb = cb
	c.arg = arg
}

func (c *CoinCrawler) crawlingProcess() {
	var flagSentMsg bool
	var cycle int

	log.Info("Coin premium notification service start")
	for !c.done {
		if !flagSentMsg && cycle == 0 {
			err := c.getPage()
			if err == nil && c.checkCoinPremium() {
				msg := c.getCoinPremium()

				// send msg to chat
				log.Debug("send msg : ", msg)
				c.cb(msg, c.arg)
				flagSentMsg = true
			}
			if err == nil {
				c.putPage()
			}
		}

		if cycle == 0 {
			cycle = 60
		} else {
			cycle--
		}

		time.Sleep(time.Second)
	}
	c.isDone <- true
}

func (c *CoinCrawler) Stop() {
	c.done = true
	<-c.isDone
	log.Info("Stop Coin premium crawling service")
}

func (c *CoinCrawler) getPage() error {
	log.Debug("Call getPage")

	c.driver = agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu", "--no-sandbox"}),
	)
	if err := c.driver.Start(); err != nil {
		log.Error(err)
		return err
	}

	page, err := c.driver.NewPage()
	if err != nil {
		log.Error(err)
		c.driver.Stop()
		return err
	}

	if err := page.Navigate("https://wisebody.co.kr"); err != nil {
		log.Error(err)
		c.driver.Stop()
		return err
	}

	c.page = page

	return nil
}

func (c *CoinCrawler) putPage() {
	log.Debug("Call putPage")

	if err := c.page.CloseWindow(); err != nil {
		log.Error(err)
	}
	if err := c.driver.Stop(); err != nil {
		log.Error(err)
	}
}

func (c *CoinCrawler) getCoinPremium() string {
	log.Debug("get coin premium")

	data := c.page.All("table.type2 > tbody > tr")
	count, _ := data.Count()

	var binance int
	var upbit int
	var premium float64

	for i := 0; i < count; i++ {
		str, _ := data.At(i).Text()
		if strings.Contains(str, "XLM") {
			binance, _ = strconv.Atoi(strings.Split(strings.Split(str, " ")[2], "\n")[1])
			upbit, _ = strconv.Atoi(strings.Split(strings.Split(str, " ")[3], "\n")[0])

			tmp := strings.Split(strings.Split(str, " ")[3], "\n")[1]
			tmp = strings.Split(tmp, "%")[0]
			premium, _ = strconv.ParseFloat(tmp, 64)

			// log.Debug("binance ", binance)
			// log.Debug("upbit ", upbit)
			// log.Debug("premium ", premium)
			break
		}
	}

	msg := fmt.Sprintf("Binance: %d, UPbit: %d (%.2f%%)", binance, upbit, premium)

	return msg
}

func (c *CoinCrawler) checkCoinPremium() bool {
	log.Debug("checkCoinPremium")

	data := c.page.All("table.type2 > tbody > tr")
	count, _ := data.Count()

	var binance int
	var upbit int
	var premium float64

	for i := 0; i < count; i++ {
		str, _ := data.At(i).Text()
		if strings.Contains(str, "XLM") {
			binance, _ = strconv.Atoi(strings.Split(strings.Split(str, " ")[2], "\n")[1])
			upbit, _ = strconv.Atoi(strings.Split(strings.Split(str, " ")[3], "\n")[0])

			tmp := strings.Split(strings.Split(str, " ")[3], "\n")[1]
			tmp = strings.Split(tmp, "%")[0]
			premium, _ = strconv.ParseFloat(tmp, 64)

			// log.Debug("binance ", binance)
			// log.Debug("upbit ", upbit)
			// log.Debug("premium ", premium)
			break
		}
	}
	log.Debugf("Binance: %d, UPbit: %d (%.2f%%)", binance, upbit, premium)

	if premium < -5.0 {
		return true
	}

	return false
}
