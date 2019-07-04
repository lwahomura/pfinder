package pfinder

import (
	"testing"
)

func TestPrequests(t *testing.T) {
	seleniumpath := "/path/to/selenium/selenium-server-standalone-x.xxx.xx.jar"
	geckopath := "/path/to/geckodriver/geckodriver-vx.x.x-system"
	login := "youremail@gmail.com"
	pw := "yourpassword"
	pr, err := CreatePRequestrer()
	if err != nil {
		t.Fatal(err)
	}
	pr.AddSources(seleniumpath, geckopath, login, pw)
	data := pr.GetProxies()
	pr.CheckProxies(data, 100)
}
