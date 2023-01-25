package handler

import (
	"net/http"
	"sync"

	"github.com/PereRohit/util/response"
)

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_common.go --package=mock github.com/vatsal278/TransactionManagementService/internal/handler Commoner,HealthChecker

type Commoner interface {
	MethodNotAllowed(http.ResponseWriter, *http.Request)
	RouteNotFound(http.ResponseWriter, *http.Request)
	HealthCheck(http.ResponseWriter, *http.Request)
}

type HealthChecker interface {
	HealthCheck() (svcName string, msg string, stat bool)
}

type common struct {
	sync.Mutex
	services []HealthChecker
}

var c = common{}

func AddHealthChecker(h HealthChecker) {
	c.Lock()
	defer c.Unlock()
	c.services = append(c.services, h)
}

func NewCommonSvc() Commoner {
	return &c
}

func (common) MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	response.ToJson(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), nil)
}

func (common) RouteNotFound(w http.ResponseWriter, _ *http.Request) {
	response.ToJson(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), nil)
}

func (common) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	type svcHealthStat struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	svcUpdate := map[string]svcHealthStat{}

	for _, svc := range c.services {
		name, msg, ok := svc.HealthCheck()
		stat := http.StatusText(http.StatusOK)
		if !ok {
			stat = "Not " + stat
		}
		svcUpdate[name] = svcHealthStat{
			Status:  stat,
			Message: msg,
		}
	}
	response.ToJson(w, http.StatusOK, http.StatusText(http.StatusOK), svcUpdate)
}
