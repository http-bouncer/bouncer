package main

import (
	"net/url"
	"regexp"
	"sort"
	"sync"
	"time"
)

type Config struct {
	Id                            int
	Host                          string
	Path                          string
	TargetPath                    string
	ReqPerSecond                  int
	MaxConcurrentPerBackendServer int
	BackendServers                []BackendServer
	NextBackendServer             chan BackendServer `json:"-"`
	quitBootstrap                 chan struct{}      `json:"-"`
	Throttle                      chan struct{}      `json:"-"`
	quitRatelimit                 chan struct{}      `json:"-"`
}

type BackendServer struct {
	Id       int
	Host     string
	ConfigId int
}

func (config *Config) BackendServerBootstrapRoutine(quit chan struct{}) {
	for i := 0; i < config.MaxConcurrentPerBackendServer; i++ {
		for _, next := range config.BackendServers {
			select {
			case <-quit:
				return
			case config.NextBackendServer <- next:
			}
		}
	}
}

func (config *Config) RateLimiterRoutine(quit chan struct{}) {
	rem := config.ReqPerSecond
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			rem = config.ReqPerSecond
		case <-quit:
			return
		case config.Throttle <- struct{}{}:
			rem--
			if rem <= 1 {
				<-ticker.C
				rem = config.ReqPerSecond
			}
		}
	}
}

func (config *Config) Reload() {
	config.Destroy()
	config.quitRatelimit = make(chan struct{}, 1)
	config.quitBootstrap = make(chan struct{}, 1)
	go config.BackendServerBootstrapRoutine(config.quitBootstrap)
	go config.RateLimiterRoutine(config.quitRatelimit)
}

func (config *Config) Destroy() {
	config.quitBootstrap <- struct{}{}
	config.quitRatelimit <- struct{}{}
}

func NewConfig(hostname string, backendServers []BackendServer, path string, targetPath string, concurrency int, reqPerSecond int) Config {
	var config Config
	config.BackendServers = backendServers
	config.NextBackendServer = make(chan BackendServer, concurrency)
	config.quitRatelimit = make(chan struct{}, 1)
	config.quitBootstrap = make(chan struct{}, 1)
	config.Throttle = make(chan struct{}, 1)
	config.MaxConcurrentPerBackendServer = concurrency
	config.ReqPerSecond = reqPerSecond
	config.Path = path
	config.TargetPath = targetPath
	config.Host = hostname
	go config.BackendServerBootstrapRoutine(config.quitBootstrap)
	go config.RateLimiterRoutine(config.quitRatelimit)
	return config
}

func NewBackendServer(url string) BackendServer {
	var backendServer BackendServer
	backendServer.Host = url
	return backendServer
}

type HostConfigs struct {
	configs []*Config
	mutex   sync.Mutex
}

func (a HostConfigs) Len() int           { return len(a.configs) }
func (a HostConfigs) Less(i, j int) bool { return len(a.configs[i].Path) > len(a.configs[j].Path) }
func (a HostConfigs) Swap(i, j int)      { a.configs[i], a.configs[j] = a.configs[j], a.configs[i] }

func (a *HostConfigs) AddConfig(config *Config) {
	a.mutex.Lock()
	a.configs = append(a.configs, config)
	sort.Sort(a)
	a.mutex.Unlock()
	go func() {
		AddConfig <- config
	}()
}

func (a *HostConfigs) RemoveConfig(config *Config) {
	a.mutex.Lock()
	for index, val := range a.configs {
		if val.Path == config.Path {
			a.configs = append(a.configs[:index], a.configs[index+1:]...)
			break
		}
	}
	sort.Sort(a)
	a.mutex.Unlock()
	go func() {
		RemoveConfig <- config
	}()
}

func (a *HostConfigs) UpdateConfig(config *Config) {
	a.mutex.Lock()
	for indx, val := range a.configs {
		if val.Path == config.Path {
			a.configs = append(a.configs[:indx], a.configs[indx+1:]...)
			a.configs = append(a.configs, config)
			break
		}
	}
	sort.Sort(a)
	a.mutex.Unlock()
	go func() {
		UpdateConfig <- config
	}()
}

func (a HostConfigs) Match(path string) *Config {
	for _, config := range a.configs {
		if match, _ := regexp.MatchString("^"+config.Path, path); match {
			return config
		}
	}
	return nil
}

type ConfigStore map[string]*HostConfigs

func (a ConfigStore) AddConfig(config *Config) {
	if _, ok := a[config.Host]; ok {
		a[config.Host].AddConfig(config)
	} else {
		var hostConfigs HostConfigs
		hostConfigs.AddConfig(config)
		a[config.Host] = &hostConfigs
	}
}

func (a ConfigStore) RemoveConfig(config *Config) {
	if val, ok := a[config.Host]; ok {
		val.RemoveConfig(config)
	}
}

func (a ConfigStore) UpdateConfig(config *Config) {
	if val, ok := a[config.Host]; ok {
		val.UpdateConfig(config)
	}
}

func (a ConfigStore) Match(target *url.URL) *Config {
	var config *Config
	if val, ok := a[target.Host]; ok {
		return val.Match(target.Path)
	}
	return config
}

func (a ConfigStore) GetAllConfigs() []*Config {
	var configs []*Config
	for _, val := range a {
		configs = append(configs, val.configs...)
	}
	return configs
}

func (a ConfigStore) GetAllHostConfigs(host string) []*Config {
	var configs []*Config
	for key, val := range a {
		if key == host {
			configs = append(configs, val.configs...)
		}
	}
	return configs
}

func (a ConfigStore) GetConfig(host, path string) *Config {
	var config *Config
	if val, ok := a[host]; ok {
		for _, conf := range val.configs {
			if conf.Path == path {
				return conf
			}
		}
	}
	return config
}
