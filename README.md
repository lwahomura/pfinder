# Package pfinder

[Overview](#overview)  
[Installing](#installing)  

## Overview
pfinder allows you to get proxies from free sources and check them.
It works the best with socks5 proxies

## Installing
````
go get github.com/lwahomura/pfinder
````
Also download selenium (https://www.seleniumhq.org/download/) and geckodriver (https://github.com/mozilla/geckodriver/releases) 
and provide paths to them to spysone service:
````
seleniumpath := "/path/to/selenium/selenium-server-standalone-x.xxx.xx.jar"
geckopath := "/path/to/geckodriver/geckodriver-vx.x.x-system"
spys := adder.SourceSpysone{}
spys.Create(seleniumpath, geckopath)
````
Also you'd better register at gatherproxy.com, because it's a nice source