package internal

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"net/http"
	"time"
	"net/url"
	"crypto/tls"
	"bytes"
	"compress/gzip"
	"io"
	"unsafe"
)

func DefaultGetRequest(target string) *http.Request {
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:62.0) Gecko/20100101 Firefox/62.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-GB,en;q=0.5")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	return req
}

func DefaultClient(proxy string) *http.Client {
	client := &http.Client{}
	timeout := time.Duration(20 * time.Second)
	client.Timeout = timeout
	if proxy != "" {
		proxy, _ := url.Parse(proxy)
		transport := http.Transport{Proxy: http.ProxyURL(proxy), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = &transport
	}
	return client
}

func GetReadableBody(resp *http.Response) string {
	var buf bytes.Buffer
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(resp.Body)
		defer reader.Close()
		io.Copy(&buf, reader)
	default:
		(&buf).ReadFrom(resp.Body)
	}

	buff := new(bytes.Buffer)
	buff.ReadFrom(&buf)
	b := buff.Bytes()
	return *(*string)(unsafe.Pointer(&b))
}

func UpdateGeobase(pauser *bool, needUpdate chan<- string,
	doneUpdate <-chan string) error {
	defer func() {
		// tell checker that the db was updated
		needUpdate <- "update"
		<-doneUpdate
		*pauser = false
	}()
	cmd := exec.Command("wget", "https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz")
	if err := cmd.Run(); err == nil {
		if err := findTar(); err == nil {
			// tell checker that db will be updated now, so connection shouldn't be used
			*pauser = true
			err := deleteMmdb()
			if err == nil ||
				regexp.MustCompile(`.* no such file or directory`).MatchString(err.Error()) {
				cmd = exec.Command("tar", "-zxvf", "GeoLite2-City.tar.gz")
				if err := cmd.Run(); err == nil {
					if err := moveMmdb(); err != nil {
						return err
					}
				}
			}
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func findTar() error {
	_, err := os.Stat("GeoLite2-City.tar.gz")
	if err != nil {
		return err
	}
	return nil
}

func deleteMmdb() error {
	err := os.Remove("GeoLite2-City.mmdb")
	if err != nil {
		return err
	}
	return nil
}

func moveMmdb() error {
	curr, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Remove("GeoLite2-City.tar.gz")
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(curr)
	if err != nil {
		return err
	}
	for _, path := range files {
		if path.IsDir() && regexp.MustCompile(`GeoLite2.*`).MatchString(path.Name()) {
			cmd := exec.Command("mv", path.Name()+"/GeoLite2-City.mmdb", ".")
			err := cmd.Run()
			if err != nil {
				return err
			}
			err = os.RemoveAll(path.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
