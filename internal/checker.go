package internal

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/lwahomura/pfinder/pkg/models"
)

type PChecker struct {
	db         *geoip2.Reader
	dbpath     string
	needUpdate <-chan string
	doneUpdate chan<- string
}

func (pc *PChecker) updateListener() {
	for {
		select {
		case mes := <-pc.needUpdate:
			if mes == "update" {
				db, err := geoip2.Open(pc.dbpath)
				if err != nil {
					log.Fatal("Couldn't create pchecker: ", err)
				}
				pc.db = db
				pc.doneUpdate <- "done"
			}
		}
	}
}

func CreatePChecker(dbpath string, needUpdate <-chan string,
	doneUpdate chan<- string) *PChecker {
	db, err := geoip2.Open(dbpath)
	if err != nil {
		if regexp.MustCompile(`no such file`).FindString(err.Error()) != "" {
			db = nil
		} else {
			log.Fatal("Couldn't create pchecker: ", err)
		}
	}
	pc := &PChecker{
		db:         db,
		dbpath:     dbpath,
		needUpdate: needUpdate,
		doneUpdate: doneUpdate,
	}

	// we listen the channel whether db is updated
	go pc.updateListener()
	return pc
}

func (pc *PChecker) CheckDbExistance() *geoip2.Reader{
	return pc.db
}

func (pc *PChecker) AddDb(db *geoip2.Reader) {
	pc.db = db
}

func (pc *PChecker) CheckProxy(proxy *models.Proxy, pauser *bool) {
	proxy.Valid = true
	// does the proxy work?
	pc.CheckConnection(proxy)
	// does it send unnecessary headers?
	pc.CheckHeaders(proxy)
	// what is it's location?

	// cab use connection unless the base is being updated
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			if !*pauser {
				pc.CheckLocation(proxy)
				return
			}
			fmt.Println("can't check")
		}
	}
}

func (pc *PChecker) CheckHeaders(proxy *models.Proxy) {
	req := DefaultGetRequest("http://request.urih.com/")
	cli := DefaultClient(proxy.ToString())
	cli.Timeout = 10 * time.Second
	resp, err := cli.Do(req)
	if err != nil || resp == nil {
		proxy.Valid = false
		return
	}
	body := GetReadableBody(resp)
	re := regexp.MustCompile(`<b>Forwarded|<b>Via|<b>X-Forwarded|<b>Connection|<b>Keep-alive|<b>Proxy|<b>TE|<b>Trailer|<b>Transfer|<b>Upgrade `)
	res := re.FindStringSubmatch(body)
	if len(res) != 0 {
		proxy.Valid = false
		return
	}
	re = regexp.MustCompile(`IP: \d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	res = re.FindStringSubmatch(body)
	if len(res) == 0 {
		proxy.Valid = false
		return
	}
	ipShort := strings.Split(res[0], " ")[1]
	if ipShort != proxy.IPAddress {
		proxy.Valid = false
		return
	}
}

func (pc *PChecker) CheckConnection(proxy *models.Proxy) {
	req := DefaultGetRequest("https://www.google.com/?q=dog")
	cli := DefaultClient(proxy.ToString())
	cli.Timeout = 10 * time.Second
	resp, err := cli.Do(req)
	if err != nil || resp == nil {
		proxy.Valid = false
		return
	}
}

func (pc *PChecker) CheckLocation(proxy *models.Proxy) {
	geoInfo, err := pc.GetGeoInfo(proxy.IPAddress)
	if err != nil {
		log.Error(err)
		proxy.Valid = false
		return
	}
	proxy.City = geoInfo.City
	proxy.Country = geoInfo.Country
	proxy.Code = geoInfo.CountryCode
	proxy.Lat = geoInfo.Lat
	proxy.Long = geoInfo.Long
	if proxy.City == "" || proxy.Country == "" || proxy.Code == "" || proxy.Lat == 0 || proxy.Long == 0 {
		// if we can't properly define proxy location - we set it undefined so that it won't bother us
		proxy.Code = ""
		proxy.Country = "undefined"
		proxy.City = "undefined"
		proxy.Lat = 0
		proxy.Long = 0
		return
	}
}

func (pc *PChecker) GetGeoInfo(ip string) (models.GeoInfo, error) {
	defer func() {
		if r := recover(); r != nil {
			if pc.db == nil {
				log.Error("Couldn't get geoinfo because there is no db")
			} else {
				log.Error("Couldn't get geoinfo because of an unexpected reason")
			}
		}
	}()
	res, err := pc.db.City(net.ParseIP(ip))
	if err != nil {
		return models.GeoInfo{}, err
	}
	return models.GeoInfo{
		City:        res.City.Names["en"],
		Country:     res.Country.Names["en"],
		CountryCode: res.Country.IsoCode,
		Lat:         res.Location.Latitude,
		Long:        res.Location.Longitude,
	}, nil
}
