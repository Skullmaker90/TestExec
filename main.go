package main

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/sd/consul"
	httptransport "github.com/go-kit/kit/transport/http"
	stdconsul "github.com/hashicorp/consul/api"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"time"
)

// Main
func main() {
	logger := log.NewJSONLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name: "request_count",
		Help: "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name: "request_latency_microseconds",
		Help: "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name: "count_result",
		Help: "The result of each count method.",
	}, []string{})

	var svc StringService
	svc = stringService{}
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	ttl := time.Second * 30
	c, err := stdconsul.NewClient(stdconsul.DefaultConfig())
	if err != nil {
		logger.Log("err", err)
	}
	client := consul.NewClient(c)
	registrar := consul.NewRegistrar(client, &stdconsul.AgentServiceRegistration{
		Name: "StringService",
		Check: &stdconsul.AgentServiceCheck{
			CheckID: "StringService",
			TTL: ttl.String(),
		},
	}, logger)
	registrar.Register()

	go func () {
		ticker := time.NewTicker(ttl / 2)
		for range ticker.C {
			agentErr := c.Agent().UpdateTTL("StringService", "Passed", "pass")
			if agentErr != nil {
				logger.Log("err", agentErr)
			}
		}
	}()

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc),
		decodeCountRequest,
		encodeResponse,
	)

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/metrics", promhttp.Handler())
	logger.Log("msg", "HTTP", "addr", ":8080")
	logger.Log("err", http.ListenAndServe(":8080", nil))
}