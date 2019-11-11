package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = fallback
	}
	fmt.Println(key, "=", value)
	return value
}

func main() {
	listenAddr := getenv("LISTEN_ADDR", ":8080")
	endpoint := getenv("ENDPOINT", "/query")
	wimp := getenv("WIMP", "false")

	errors := make(chan error)
	handler := func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			errors <- err
			return
		}
		response := fmt.Sprintf("%v: %s", time.Now(), body)
		response = fmt.Sprintf("%x\n", md5.Sum([]byte(response)))
		io.WriteString(w, response)
		if strings.EqualFold(wimp, "true") {
			close(errors)
		}
	}

	http.HandleFunc(endpoint, handler)
	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(255)
	}()

	for err := range errors {
		if err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	fmt.Println("Finished")
}
