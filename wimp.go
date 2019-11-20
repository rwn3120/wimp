package main

import (
	"crypto/md5"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
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
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Bad Request:", err.Error())
		http.Error(w, "can't read body", http.StatusBadRequest)
		done <- true
		return
	}
	duration := time.Duration(rand.Int63n(maxDuration.Milliseconds()-minDuration.Milliseconds())+minDuration.Milliseconds()) * time.Millisecond
	<-time.After(duration)
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf("%v: %s", time.Now(), body)
	response = fmt.Sprintf("%x", md5.Sum([]byte(response)))
	fmt.Println("OK:", response)
	io.WriteString(w, response+"\n")
	done <- true
}

func main() {
	listenAddr := getEnv("LISTEN_ADDR", ":8080")
	queryEndpoint := getEnv("QUERY_ENDPOINT", "/query")
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
