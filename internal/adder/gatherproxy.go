package adder

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/lwahomura/pfinder/internal"
	"github.com/lwahomura/pfinder/pkg/models"
)

type SourceGatherproxy struct {
	login string
	pw    string
}

func (sg *SourceGatherproxy) Create(opts ...interface{}) {
	if len(opts) < 2 || reflect.TypeOf(opts[0]).String() != "string" ||
		reflect.TypeOf(opts[1]).String() != "string" {
		log.Fatal("Couldn't create gatherproxy source because of wrong" +
			" provided creds")
	}
	sg.login = opts[0].(string)
	sg.pw = opts[1].(string)
}

func (sg *SourceGatherproxy) GetProxylist() string {
	client := internal.DefaultClient("")
	cookies, ans, err := sg.firstVisitGatherproxy(client)
	if err != nil {
		log.Error("Counldn't get proxies from gatherproxy: ", err)
		return ""
	}
	err = sg.loginProxylist(client, cookies, ans)
	if err != nil {
		log.Error("Counldn't get proxies from gatherproxy: ", err)
		return ""
	}
	data, err := sg.downloadProxylist(client, cookies)
	if err != nil {
		log.Error("Counldn't get proxies from gatherproxy: ", err)
		return ""
	}
	return data
}

func (sg *SourceGatherproxy) GetProxyStrings(data string) []string {
	return strings.Split(data, "\r\n")
}

func (sg *SourceGatherproxy) ConvertStrings(data []string) []*models.Proxy {
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
	fmt.Println("gather done")
	return res
}

// open login page
func (sg *SourceGatherproxy) firstVisitGatherproxy(client *http.Client) (
	[]*http.Cookie, int, error) {
	req := internal.DefaultGetRequest("http://www.gatherproxy.com/subscribe/login")
	resp, err := client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	cookies := resp.Cookies()
	body := internal.GetReadableBody(resp)

	// we need to solve captcha and send it's result with login request
	task, err := sg.parseProxylistBodyTask(body)
	if err != nil {
		return nil, -1, err
	}
	ans, err := sg.parseProxylistTask(task)
	if err != nil {
		return nil, -1, err
	}

	return cookies, ans, nil
}


func (sg *SourceGatherproxy) parseProxylistBodyTask(stringStream string) (string, error) {
	var task string
	taskString := regexp.MustCompile(`Enter verify code: <span [^>]+>[^<]+</span>`).FindString(stringStream)
	if taskString == "" {
		return task, errors.New("no task part")
	}
	parts := strings.Split(taskString, ">")
	task = regexp.MustCompile(`[^<]+`).FindString(parts[1])
	return task, nil
}

func (sg *SourceGatherproxy) parseProxylistTask(task string) (int, error) {
	var parsedTask string
	var answer int
	task = strings.ToLower(task)
	task = regexp.MustCompile(`[^=]+`).FindString(task)
	parts := strings.Split(task, " ")
	for _, item := range parts {
		switch item {
		case "1", "one":
			parsedTask += "1"
		case "2", "two":
			parsedTask += "2"
		case "3", "three":
			parsedTask += "3"
		case "4", "four":
			parsedTask += "4"
		case "5", "five":
			parsedTask += "5"
		case "6", "six":
			parsedTask += "6"
		case "7", "seven":
			parsedTask += "7"
		case "8", "eight":
			parsedTask += "8"
		case "9", "nine":
			parsedTask += "9"
		case "0", "zero":
			parsedTask += "0"
		case "+", "plus":
			parsedTask += "+"
		case "-", "minus":
			parsedTask += "-"
		case "x", "multiplied":
			parsedTask += "*"
		case "\\", "divided":
			parsedTask += "\\"
		}
	}
	if len(parsedTask) != 3 {
		return answer, errors.New("couldn't resolve task")
	}
	first, _ := strconv.Atoi(parsedTask[0:1])
	second, _ := strconv.Atoi(parsedTask[2:3])
	switch parsedTask[1:2] {
	case "+":
		answer = first + second
	case "-":
		answer = first - second
	case "*":
		answer = first * second
	case "\\":
		answer = first / second

	}
	return answer, nil
}

func (sg *SourceGatherproxy) loginProxylist(client *http.Client, cookies []*http.Cookie, taskAns int) error {
	url := "http://www.gatherproxy.com/subscribe/login"
	ans := strconv.Itoa(taskAns)

	query := fmt.Sprintf("Username=%s&Password=%s&Captcha=%s&undefined=",
		sg.login, sg.pw, ans)
	payload := strings.NewReader(query)

	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	for _, item := range cookies {
		req.AddCookie(item)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (sg *SourceGatherproxy) downloadProxylist(client *http.Client,
	cookies []*http.Cookie) (string, error) {
	req, _ := http.NewRequest("POST", "http://www.gatherproxy.com/sockslist/plaintext", nil)
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, item := range cookies {
		req.AddCookie(item)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	return internal.GetReadableBody(resp), err
}
