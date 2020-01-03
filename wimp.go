package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = fallback
	}
	fmt.Println(key, "=", value)
	return value
}

func getEnvDuration(key, fallback string) time.Duration {
	duration, err := time.ParseDuration(getEnv(key, fallback))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	}
	return duration
}

func serve(w http.ResponseWriter, req *http.Request, minDuration, maxDuration time.Duration, done chan bool) {
	duration := time.Duration(rand.Int63n(maxDuration.Milliseconds()-minDuration.Milliseconds())+minDuration.Milliseconds()) * time.Millisecond
	<-time.After(duration)
	w.WriteHeader(http.StatusOK)
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	response := fmt.Sprintf("Time %v\n---\n%s\n---\n%s\n", time.Now().Format(time.ANSIC), strings.Join(os.Environ(), "\n"), dump)
	fmt.Println(response)
	io.WriteString(w, response)
	done <- true
}

func main() {
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	queryEndpoint := getEnv("QUERY_ENDPOINT", "/")
	metricsEndpoint := getEnv("METRICS_ENDPOINT", "/metrics")
	minDuration := getEnvDuration("MIN_QUERY_DURATIOn", "50ms")
	maxDuration := getEnvDuration("MAX_QUERY_DURATION", "500ms")
	ttl := getEnvDuration("TTL", "87600h") // 10y

	ready := make(chan bool, 1)
	defer close(ready)
	ready <- true
	queryHandler := func(w http.ResponseWriter, req *http.Request) {
		Total.Inc()
		select {
		case <-ready:
			Accepted.Inc()
			Status.Inc()
			serve(w, req, minDuration, maxDuration, ready)
			Status.Dec()
		default:
			Refused.Inc()
			http.Error(w, "Already processing some query", http.StatusConflict)
		}
	}

	http.HandleFunc(queryEndpoint, queryHandler)
	http.Handle(metricsEndpoint, promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(255)
	}()
	<-time.After(ttl)
	fmt.Println("Died")
}
