package adder

import (
	"strings"
	"github.com/lwahomura/pfinder/internal"
	"github.com/lwahomura/pfinder/pkg/models"
	"fmt"
)

type SourceSocksproxy struct {

}

func (ss *SourceSocksproxy) Create(opts... interface{}) {

}

func (ss *SourceSocksproxy) GetProxylist() string {
	req := internal.DefaultGetRequest("https://socks-proxy.net/")
	cli := internal.DefaultClient("")
	resp, err := cli.Do(req)
	if err != nil || resp == nil {
		return ""
	}
	body := internal.GetReadableBody(resp)
	return body
}

func (ss *SourceSocksproxy) GetProxyStrings(data string) []string {
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

func (ss *SourceSocksproxy) ConvertStrings(data []string) []*models.Proxy {
	var res []*models.Proxy
	for _, item := range data {
		var lines []string
		item = strings.Replace(item, "</td><", "</td>\n<", -1)
		lines = strings.Split(item, "\n")
		if len(lines) == 8 {
			proxy := models.Proxy{}
			proxy.CreateByArr(lines)
			res = append(res, &proxy)
		}
	}
	fmt.Println("socks done")
	return res
}