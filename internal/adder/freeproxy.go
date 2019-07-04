package adder

import (
	"encoding/base64"
	"regexp"
	"strings"
	"github.com/lwahomura/pfinder/internal"
	"github.com/lwahomura/pfinder/pkg/models"
	"fmt"
	"strconv"
)

type SourceFreeproxy struct {

}

func (sf *SourceFreeproxy) Create(opts... interface{}) {

}

func (sf *SourceFreeproxy) GetProxylist() string {
	res := ""
	url := "http://free-proxy.cz/en/proxylist/country/US/socks5/ping/all/"
	i := 0
	for {
		i++
		req := internal.DefaultGetRequest(url + strconv.Itoa(i))
		cli := internal.DefaultClient("")
		resp, err := cli.Do(req)
		if err != nil || resp == nil || !strings.Contains(resp.Status, "200") {
			break
		}
		body := internal.GetReadableBody(resp)
		res += body + "\n"
	}
	return res
}

func (sf *SourceFreeproxy) GetProxyStrings(data string) []string {
	var testStr []string
	var proxyPart string
	var cells []string
	data = strings.Replace(data, "<tbody>", "CUT_HERE<tbody>", -1)
	data = strings.Replace(data, "</tbody>", "</tbody>CUT_HERE", -1)
	testStr = strings.Split(data, "CUT_HERE")
	for _, item := range testStr {
		if strings.Contains(item, "<tbody>") {
			proxyPart = item
		}
	}
	proxyPart = strings.Replace(proxyPart, "<tbody>", "", -1)
	proxyPart = strings.Replace(proxyPart, "</tbody>", "", -1)
	proxyPart = strings.Replace(proxyPart, "<tr>", "", -1)
	cells = strings.Split(proxyPart, "</tr>")
	return cells[:len(cells)-1]
}

func (sf *SourceFreeproxy) ConvertStrings(data []string) []*models.Proxy {
	var res []*models.Proxy
	for _, item := range data {
		if strings.Contains(item, "colspan=\"11\"") == false {
			proxy := models.Proxy{}
			re := regexp.MustCompile(`Base64\.decode\(\".*\"\)`)

			ipaddr := re.FindStringSubmatch(item)[0]
			re = regexp.MustCompile(`\".*\"`)
			ipaddr = re.FindStringSubmatch(ipaddr)[0]
			data, _ := base64.StdEncoding.DecodeString(ipaddr[1 : len(ipaddr)-1])
			ipaddr = string(data)

			re = regexp.MustCompile(`style=\'\'\>\d*\<`)
			port := re.FindStringSubmatch(item)[0]
			re = regexp.MustCompile(`\d{1,5}`)
			port = re.FindStringSubmatch(port)[0]
			proxy.Valid = true
			proxy.IPAddress = ipaddr
			proxy.Port = port
			proxy.ProxyType = "socks5"
			res = append(res, &proxy)
		}
	}
	fmt.Println("free done")
	return res
}
