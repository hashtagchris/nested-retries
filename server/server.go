package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashtagchris/nested-retries/client"
)

type server struct {
	addr           string
	log            *log.Logger
	nextServerPort int
	errResp        bool
}

func NewIntermediateServer(port, nextServerPort int) *server {
	return newServer(port, nextServerPort, false)
}

func NewTerminalServer(port int, errResp bool) *server {
	return newServer(port, 0, errResp)
}

func newServer(port, nextServerPort int, errResp bool) *server {
	addr := fmt.Sprintf(":%d", port)
	logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", addr), log.Ltime)

	return &server{addr, logger, nextServerPort, errResp}
}

func (s *server) Run() {
	hs := &http.Server{
		Addr:    s.addr,
		Handler: s,
	}

	s.log.Fatal(hs.ListenAndServe())
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.Println("received request")

	var depth int64
	if s.nextServerPort > 0 {
		serverDepth, err := client.GetDepth(s.nextServerPort)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		depth = serverDepth + 1
	} else if s.errResp {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		depth = 1
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, depth)
}
