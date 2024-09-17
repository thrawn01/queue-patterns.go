package queue

import (
	"fmt"
	"github.com/duh-rpc/duh-go"
	v1 "github.com/duh-rpc/duh-go/proto/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thrawn01/queue-patterns.go/proto"
	"net/http"
	"time"
)

type HTTPHandler struct {
	duration *prometheus.SummaryVec
	log      duh.StandardLogger
	metrics  http.Handler
}

func NewHTTPHandler(metrics http.Handler, log duh.StandardLogger) *HTTPHandler {
	return &HTTPHandler{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "http_handler_duration",
			Help: "The timings of http requests handled by the service",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.99: 0.001,
			},
		}, []string{"path"}),
		metrics: metrics,
		log:     log,
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer prometheus.NewTimer(h.duration.WithLabelValues(r.URL.Path)).ObserveDuration()

	if r.URL.Path == "/metrics" && r.Method == http.MethodGet {
		h.metrics.ServeHTTP(w, r)
		return
	}

	if r.Method != http.MethodPost {
		duh.ReplyWithCode(w, r, duh.CodeBadRequest, nil,
			fmt.Sprintf("http method '%s' not allowed; only POST", r.Method))
		return
	}

	var req proto.ProduceRequest
	if err := duh.ReadRequest(r, &req, duh.MegaByte*50); err != nil {
		duh.ReplyWithCode(w, r, duh.CodeInternalError, nil, err.Error())
		return
	}

	// Pretend to do some work that takes 10 Milliseconds
	time.Sleep(10 * time.Millisecond)

	duh.Reply(w, r, duh.CodeOK, &v1.Reply{Code: duh.CodeOK})
}

// Describe fetches prometheus metrics to be registered
func (h *HTTPHandler) Describe(ch chan<- *prometheus.Desc) {
	h.duration.Describe(ch)
}

// Collect fetches metrics from the server for use by prometheus
func (h *HTTPHandler) Collect(ch chan<- prometheus.Metric) {
	h.duration.Collect(ch)
}
