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
	config := NewConfig(
		"localhost:9090",
		[]BackendServer{NewBackendServer("localhost:9091"), NewBackendServer("localhost:9092")},
		"/",
		"/",
		10,
		200)
	defaultConfig = &config
	configStore.AddConfig(defaultConfig)
	configCurrentStats = make(map[*Config]*CurrentStats)
	globalStatSubscribers.Init()
	configStatSubscribers.Init()
	go statProcessor()
	go GlobalStatBroadcaster()
	go reqsPrinter()
	go UIServer()
	defaultConfig.MaxConcurrentPerBackendServer = 20
	defaultConfig.ReqPerSecond = 40
	defaultConfig.Reload()
	http.ListenAndServe(":9090", proxy)
}

func director(req *http.Request) (*Config, *BackendServer) {
	config := configStore.Match(req.Host, req.URL.Path)
	<-config.Throttle
	select {
	case next := <-config.NextBackendServer:
		req.URL.Scheme = "http"
		req.URL.Host = next.Host
		return config, &next
	case <-time.After(120 * time.Second):
		return nil, nil

	}
}
