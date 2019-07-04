package models

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"strings"
)

type Proxy struct {
	Id        int     `json:"id"`
	IPAddress string  `json:"ipaddress"`
	Port      string  `json:"port"`
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	ProxyType string  `json:"proxytype"`
	Code      string  `json:"countrycode"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Lat       float64 `json:"lat"`
	Long      float64 `json:"long"`
	Trusted   bool    `json:"trusted"`
	Valid     bool    `json:"valid"`
}

func (p *Proxy) CreateByArr(data []string) {
	p.Valid = true
	p.ParseIPAddress(data[0])
	p.ParsePort(data[1])
	p.ParseCode(data[2])
	p.ParseCountry(data[3])
	p.ParseVersion(strings.ToLower(data[4]))
	p.ParseAnonymity(data[5])
	p.ParseHttps(data[6])
	p.ParseLastChecked(data[7])
}

func (p *Proxy) ToString() string {
	if p.IPAddress != "" && p.Port != "" {
		if p.Username != "" && p.Password != "" {
			return fmt.Sprintf("%s://%s:%s@%s:%s", p.ProxyType, p.Username, p.Password, p.IPAddress, p.Port)
		} else {
			return fmt.Sprintf("%s://%s:%s", p.ProxyType, p.IPAddress, p.Port)
		}
	} else {
		return ""
	}
}

func (p *Proxy) ParseIPAddress(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td>`)
		content = re.ReplaceAllString(line, "")
		if content != "" && checkIP(content) {
			p.IPAddress = content
		} else {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParsePort(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td>`)
		content = re.ReplaceAllString(line, "")
		if content != "" && checkPort(content) {
			p.Port = content
		} else {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseCode(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td>`)
		content = re.ReplaceAllString(line, "")
		if checkCode(content) {
			p.Code = content
		} else {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseCountry(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td( class='hm')?>`)
		content = re.ReplaceAllString(line, "")
		if content != "" && checkCountry(content) {
			p.Country = content
		} else {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseVersion(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td>`)
		content = re.ReplaceAllString(line, "")
		if content != "" {
			p.ProxyType = content
		} else {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseAnonymity(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td( class='hm')?>`)
		content = re.ReplaceAllString(line, "")
		if checkAnonymity(content) == false {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseHttps(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td( class='hm')?>`)
		content = re.ReplaceAllString(line, "")
		if checkHttps(content) == false {
			p.Valid = false
		}
	}
}

func (p *Proxy) ParseLastChecked(line string) {
	if p.Valid {
		var content string
		re := regexp.MustCompile(`</?td( class='hd')?>`)
		content = re.ReplaceAllString(line, "")
		if checkLastChecked(content) == false {
			p.Valid = false
		}
	}
}

func checkIP(ipaddress string) bool {
	return net.ParseIP(ipaddress) != nil
}

func checkPort(port string) bool {
	return port >= "1" && port <= string(int(math.Pow(2, 16))-1)
}

func checkCode(code string) bool {
	re := regexp.MustCompile(`[A-Z]{2}`)
	return re.MatchString(code)
}

func checkCountry(country string) bool {
	re := regexp.MustCompile(`(\w*(\s|,)?)*`)
	return re.MatchString(country)
}

func checkAnonymity(anonymity string) bool {
	return anonymity == "Anonymous"
}

func checkHttps(https string) bool {
	return https == "Yes"
}
func checkLastChecked(checked string) bool {
	re := regexp.MustCompile(`[0-9]{1,2} (minute|second(s)?) ago`)
	return re.MatchString(checked)
}
