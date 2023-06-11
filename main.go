package main

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"syscall"

	"github.com/apexskier/strava-tile-proxy/service"
)

var logger = log.Default()

func errorMiddleware(h func(rw http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if err := h(rw, r); err != nil {
			if errors.Is(err, syscall.EPIPE) {
				// ignore broken pipe errors, client cancelled connection
				return
			}
			logger.Printf("error: %s, %v", r.URL.String(), err)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func main() {
	s, err := service.New()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/personal/", errorMiddleware(s.ServePersonalTile))
	mux.Handle("/global/", errorMiddleware(s.ServeGlobalTile))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Wait()
}
