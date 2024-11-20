package efincore

import "sync"

const (
	StatActiveConnections      string = "active-connections"
	StatActiveConnectRequests  string = "active-connect-requests"
	StatActiveUpgradedRequests string = "active-upgraded-requests"
	StatInterceptedRequests    string = "intercepted-requests"
	StatInterceptedResponses   string = "intercepted-responses"
	StatUpgradedRequests       string = "upgraded-requests"
)

type StatsService interface {
	Increase(stat string)
	Decrease(stat string)
	Set(stat string, value int)
	Get() Stats
	Reset()
}

type Stats map[string]int

type defaultStatsService struct {
	mutex *sync.Mutex
	stats Stats
}

var statsService StatsService

func init() {
	statsService = newStatsService()
}

func newStatsService() StatsService {
	return &defaultStatsService{
		mutex: &sync.Mutex{},
		stats: Stats{
			StatActiveConnections:      0,
			StatActiveConnectRequests:  0,
			StatActiveUpgradedRequests: 0,
			StatInterceptedRequests:    0,
			StatInterceptedResponses:   0,
			StatUpgradedRequests:       0,
		},
	}
}

func GetStatsService() StatsService {
	return statsService
}

func (ss *defaultStatsService) Increase(stat string) {
	if ss == nil {
		return
	}

	ss.mutex.Lock()
	ss.stats[stat]++
	ss.mutex.Unlock()
}

func (ss *defaultStatsService) Decrease(stat string) {
	if ss == nil {
		return
	}

	ss.mutex.Lock()
	ss.stats[stat]--
	ss.mutex.Unlock()
}

func (ss *defaultStatsService) Set(stat string, value int) {
	if ss == nil {
		return
	}

	ss.mutex.Lock()
	ss.stats[stat] = value
	ss.mutex.Unlock()
}

func (ss *defaultStatsService) Get() Stats {
	if ss == nil {
		return Stats{}
	}

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	result := Stats{}
	for k, v := range ss.stats {
		result[k] = v
	}

	return result
}

func (ss *defaultStatsService) Reset() {
	if ss == nil {
		return
	}

	ss.mutex.Lock()
	defer ss.mutex.Lock()

	newStats := map[string]int{}
	for k, _ := range ss.stats {
		newStats[k] = 0
	}
	ss.stats = newStats
}
