package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type CurrentStats struct {
	sync.RWMutex
	Hits           int64
	TotalHits      int64
	LastCalculated time.Time
	ResponseTimes  []time.Duration
	ResponseCodes  []int
}

type GlobalStatRecord struct {
	Config               *Config
	StartTime            time.Time
	EndTime              time.Time
	AverageResponseTime  time.Duration
	ResponseStatusCounts map[int]int
	TotalRequests        int64
}

var globalCurrentStats CurrentStats
var configCurrentStats map[*Config]*CurrentStats
var globalStatSink chan GlobalStatRecord

func recordStat(config *Config, res *http.Response, elapsed time.Duration) {
	_, exists := configCurrentStats[config]
	if !exists {
		configCurrentStats[config] = &CurrentStats{}
	}
	addToStat(&globalCurrentStats, res, elapsed)
	addToStat(configCurrentStats[config], res, elapsed)
}

func addToStat(stat *CurrentStats, res *http.Response, elapsed time.Duration) {
	stat.Lock()
	stat.Hits++
	stat.TotalHits++
	stat.ResponseTimes = append(stat.ResponseTimes, elapsed)
	stat.ResponseCodes = append(stat.ResponseCodes, res.StatusCode)
	stat.Unlock()
}

func processStats() {
	globalStatSink <- processStat(nil, &globalCurrentStats)
	for config, stat := range configCurrentStats {
		globalStatSink <- processStat(config, stat)
	}
}

func processStat(config *Config, currentStat *CurrentStats) GlobalStatRecord {
	var statRecord GlobalStatRecord
	statRecord.Config = config

	currentStat.Lock()
	statRecord.TotalRequests = currentStat.Hits
	statRecord.StartTime = currentStat.LastCalculated
	statRecord.EndTime = time.Now()

	currentStat.Hits = 0
	currentStat.LastCalculated = statRecord.EndTime
	responseTimes := currentStat.ResponseTimes
	responseCodes := currentStat.ResponseCodes
	currentStat.ResponseTimes = []time.Duration{}
	currentStat.ResponseCodes = []int{}
	currentStat.Unlock()

	statRecord.AverageResponseTime = GetAverageDuration(responseTimes)
	statRecord.ResponseStatusCounts = MakeFrequencyMap(responseCodes)
	return statRecord
}

func statProcessor() {
	for {
		processStats()
		time.Sleep(2 * time.Second)
	}
}

type GlobalStatSubscribers struct {
	data []chan GlobalStatRecord
	sync.Mutex
}

func (subscribers *GlobalStatSubscribers) GetSubscribers() []chan GlobalStatRecord {
	subscribers.Lock()
	defer subscribers.Unlock()
	return subscribers.data
}

func (subscribers *GlobalStatSubscribers) Init() {
	subscribers.data = make([]chan GlobalStatRecord, 1)
}

func (subscribers *GlobalStatSubscribers) Add(c chan GlobalStatRecord) {
	subscribers.Lock()
	defer subscribers.Unlock()
	subscribers.data = append(subscribers.data, c)
}

func (subscribers *GlobalStatSubscribers) Remove(c chan GlobalStatRecord) {
	subscribers.Lock()
	defer subscribers.Unlock()
	index := -1
	for i, sc := range subscribers.data {
		if sc == c {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	subscribers.data = append(subscribers.data[:index], subscribers.data[index+1:]...)
}

var globalStatSubscribers GlobalStatSubscribers

type ConfigStatSubscribers struct {
	data map[*Config][]chan GlobalStatRecord
	sync.Mutex
}

var configStatSubscribers ConfigStatSubscribers

func (subscribers *ConfigStatSubscribers) GetSubscribers(config *Config) []chan GlobalStatRecord {
	subscribers.Lock()
	defer subscribers.Unlock()
	return subscribers.data[config]
}

func (subscribers *ConfigStatSubscribers) Init() {
	subscribers.data = make(map[*Config][]chan GlobalStatRecord)
}

func (subscribers *ConfigStatSubscribers) Add(config *Config, c chan GlobalStatRecord) {
	subscribers.Lock()
	defer subscribers.Unlock()
	_, exists := subscribers.data[config]
	if !exists {
		subscribers.data[config] = make([]chan GlobalStatRecord, 1)
	}
	subscribers.data[config] = append(subscribers.data[config], c)
}

func (subscribers *ConfigStatSubscribers) Remove(config *Config, c chan GlobalStatRecord) {
	subscribers.Lock()
	defer subscribers.Unlock()
	index := -1
	for i, sc := range subscribers.data[config] {
		if sc == c {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	subscribers.data[config] = append(subscribers.data[config][:index], subscribers.data[config][index+1:]...)
}

func SubscribeGlobalStats(c chan GlobalStatRecord) {
	globalStatSubscribers.Add(c)
}

func UnsubscribeGlobalStats(c chan GlobalStatRecord) {
	globalStatSubscribers.Remove(c)
}

func SubscribeConfigStats(config *Config, c chan GlobalStatRecord) {
	configStatSubscribers.Add(config, c)
}

func UnsubscribeConfigStats(config *Config, c chan GlobalStatRecord) {
	configStatSubscribers.Remove(config, c)
}

func GlobalStatBroadcaster() {
	for {
		astat := <-globalStatSink
		if astat.Config == nil {
			for _, c := range globalStatSubscribers.GetSubscribers() {
				go func() {
					c <- astat
				}()
			}
		} else {
			for _, c := range configStatSubscribers.GetSubscribers(astat.Config) {
				go func() {
					log.Println("X", astat.Config.Path, astat)
					c <- astat
				}()
			}
		}
	}
}

func reqsPrinter() {
	c := make(chan GlobalStatRecord)
	nc := make(chan GlobalStatRecord)
	SubscribeGlobalStats(c)
	SubscribeConfigStats(defaultConfig, nc)
	for {
		select {
		case astat := <-c:
			log.Println(astat)
		case astat := <-nc:
			log.Println("CONFIG", astat)
		}
	}
}
