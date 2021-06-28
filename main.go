package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/apexskier/strava-tile-proxy/service"
)

var logger = log.Default()

func main() {
	s, err := service.New()
	if err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if err := s.ServeTile(rw, r); err != nil {
			logger.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		}
	})); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Wait()
}
