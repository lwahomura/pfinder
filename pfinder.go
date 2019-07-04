package pfinder

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"github.com/lwahomura/pfinder/internal"
	"github.com/oschwald/geoip2-golang"
	"github.com/lwahomura/pfinder/internal/adder"
	"github.com/lwahomura/pfinder/pkg/models"
)

// pauser is necessary to not let checker use db connection when the base is being updated
type PRequester struct {
	geodbname  string
	pauser     *bool
	needUpdate chan string
	doneUpdate chan string
	Sources    []models.ProxySource
	Checker    internal.PChecker
}

func CreatePRequestrer() (*PRequester, error) {
	pr := &PRequester{
		pauser:     new(bool),
		geodbname:  "./GeoLite2-City.mmdb",
		needUpdate: make(chan string, 1),
		doneUpdate: make(chan string, 1),
	}
	pr.Checker = *pr.addChecker()
	return pr, nil
}

func (pr *PRequester) GetChecker() *internal.PChecker {
	return &pr.Checker
}

func (pr *PRequester) addChecker() *internal.PChecker {
	checker := internal.CreatePChecker(pr.geodbname, pr.needUpdate, pr.doneUpdate)
	if checker.CheckDbExistance() == nil {
		if err := pr.UpdateGeobase(); err != nil {
			log.Fatal("Couldn't update db")
		}
		db, err := geoip2.Open(pr.geodbname)
		if err != nil {
			log.Fatal("Couldn't open db")
		}
		checker.AddDb(db)
	}
	return checker
}

func (pr *PRequester) UpdateGeobase() error {
	if err := internal.UpdateGeobase(pr.pauser, pr.needUpdate,
		pr.doneUpdate); err != nil {
		return err
	}
	return nil
}

// seleniumpath, geckopath - paths to binaries to use selenium;
// gplogin, gppassword - login and password for gatherproxy;
func (pr *PRequester) AddSources(seleniumpath, geckopath, gplogin, gppassword string) {
	spys := adder.SourceSpysone{}
	spys.Create(seleniumpath, geckopath)
	gather := adder.SourceGatherproxy{}
	gather.Create(gplogin, gppassword)
	socks := adder.SourceSocksproxy{}
	free := adder.SourceFreeproxy{}
	pr.Sources = append(pr.Sources,
		&spys,
		&gather,
		&socks,
		&free,
	)
}

func (pr *PRequester) GetProxies() []*models.Proxy {
	var res []*models.Proxy
	wg := &sync.WaitGroup{}
	resChan := make(chan []*models.Proxy, len(pr.Sources))
	for _, source := range pr.Sources {
		wg.Add(1)
		go func(resChan chan<- []*models.Proxy, source models.ProxySource,
			wg *sync.WaitGroup) {
			defer wg.Done()
			list := source.GetProxylist()
			arr := source.GetProxyStrings(list)
			resChan <- source.ConvertStrings(arr)
		}(resChan, source, wg)
	}
	wg.Wait()
	close(resChan)
	for r := range resChan {
		res = append(res, r...)
	}
	return res
}

func (pr *PRequester) checker(proxies <-chan *models.Proxy,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for proxy := range proxies {
		pr.Checker.CheckProxy(proxy, pr.pauser)
	}
}

// threads - count of workers(goroutines) handling proxies
func (pr *PRequester) CheckProxies(data []*models.Proxy, threads int) {
	wg := &sync.WaitGroup{}
	proxies := make(chan *models.Proxy, len(data))
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go pr.checker(proxies, wg)
	}
	for _, item := range data {
		proxies <- item
	}
	close(proxies)
	wg.Wait()
}
