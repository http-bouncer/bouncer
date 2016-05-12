package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

var defaultConfig *Config
var configStore ConfigStore

func main() {
	configStore = make(ConfigStore)
	globalStatSink = make(chan GlobalStatRecord, 10)
	proxy := &ReverseProxy{Director: director}
	configCurrentStats = make(map[*Config]*CurrentStats)
	globalStatSubscribers.Init()
	configStatSubscribers.Init()
	LoadDatabase()
	go DbOperaitons()
	go statProcessor()
	go GlobalStatBroadcaster()
	go reqsPrinter()
	go UIServer()
	http.ListenAndServe(":9090", proxy)
}

func director(req *http.Request) (*Config, *BackendServer) {
	config := configStore.Match(req.Host, req.URL.Path)
	<-config.Throttle
	select {
	case next := <-config.NextBackendServer:
		req.URL.Scheme = "http"
		req.URL.Host = next.Host
		req.URL.Path = singleJoiningSlash(config.TargetPath, req.URL.Path)
		return config, &next
	case <-time.After(120 * time.Second):
		return nil, nil

	}
}
