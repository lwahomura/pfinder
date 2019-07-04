package adder

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"

	"regexp"
	"strings"
	"reflect"
	"github.com/lwahomura/pfinder/pkg/models"
)

type SourceSpysone struct {
	wd      selenium.WebDriver
	service *selenium.Service
}

func (ss *SourceSpysone) Create(opts ... interface{}) {
	if len(opts) < 2 || reflect.TypeOf(opts[0]).String() != "string" ||
		reflect.TypeOf(opts[1]).String() != "string" {
		log.Fatal("Couldn't create spysone source because of wrong" +
			" provided creds")
	}
	seleniumPath := opts[0].(string)
	geckoPath := opts[1].(string)
	if seleniumPath == "" || geckoPath == "" {
		return
	}
	port := 8081

	selOpts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.GeckoDriver(geckoPath),
	}

	service, err := selenium.NewSeleniumService(seleniumPath, port, selOpts...)
	if err != nil {
		log.Fatal("Couldn't create spysone service: ", err)
	}
	ss.service = service

	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Fatal("Couldn't create spysone service: ", err)
	}
	ss.wd = wd
}

func (ss *SourceSpysone) GetProxylist() string {
	if ss.wd == nil {
		return ""
	}
	var res string
	if err := ss.wd.Get("http://spys.one/proxies/"); err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}

	num, err := ss.wd.FindElement(selenium.ByID, "xpp")
	if err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}
	values, err := num.FindElements(selenium.ByTagName, "option")
	if err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}
	for _, item := range values {
		val, err := item.GetAttribute("value")
		if err != nil {
			log.Error("Counldn't get proxies from spysone: ", err)
			return res
		}
		if val == "5" {
			if err := item.Click(); err != nil {
				log.Error("Counldn't get proxies from spysone: ", err)
				return res
			}
		}
	}

	typ, err := ss.wd.FindElement(selenium.ByID, "xf5")
	if err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}
	values, err = typ.FindElements(selenium.ByTagName, "option")
	if err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}
	for _, item := range values {
		val, err := item.GetAttribute("value")
		if err != nil {
			log.Error("Counldn't get proxies from spysone: ", err)
			return res
		}
		if val == "2" {
			if err := item.Click(); err != nil {
				log.Error("Counldn't get proxies from spysone: ", err)
				return res
			}
		}
	}

	proxies, err := ss.wd.FindElements(selenium.ByCSSSelector, ".spy14")
	if err != nil {
		log.Error("Counldn't get proxies from spysone: ", err)
		return res
	}
	for _, proxy := range proxies {
		text, err := proxy.Text()
		if err == nil && regexp.MustCompile(`(\d+\.?){4}:\d+`).MatchString(text) {
			res += text + "\n"
		}
	}
	return res
}

func (ss *SourceSpysone) GetProxyStrings(data string) []string {
	return strings.Split(data, "\n")
}

func (ss *SourceSpysone) ConvertStrings(data []string) []*models.Proxy {
	var res []*models.Proxy
	for _, item := range data {
		proxydata := strings.Split(item, ":")
		if len(proxydata) > 1 {
			proxy := models.Proxy{
				IPAddress: proxydata[0],
				Port:      proxydata[1],
				Valid:     true,
				ProxyType: "socks5",
			}
			res = append(res, &proxy)
		}
	}
	fmt.Println("spys done")
	return res
}
