package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/hashtagchris/nested-retries/client"
)

type Server interface {
	ID() string
	Run()
	RequestCount() int
	Reset()
}

type server struct {
	addr           string
	log            *log.Logger
	nextServerPort int
	// return 500 Internal Server responses
	serverErrResp bool
	// let the request timeout to simulate a server under extreme load
	reqTimeout bool
	mu         sync.Mutex
	count      int
}

func NewIntermediateServer(port, nextServerPort int) Server {
	return newServer(port, nextServerPort, false, false)
}

func NewTerminalServer(port int, serverErrResp, reqTimeout bool) Server {
	return newServer(port, 0, serverErrResp, reqTimeout)
}

func newServer(port, nextServerPort int, serverErrResp, reqTimeout bool) *server {
	addr := fmt.Sprintf(":%d", port)
	logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", addr), log.Ltime)

	return &server{addr, logger, nextServerPort, serverErrResp, reqTimeout, sync.Mutex{}, 0}
}

func (s *server) ID() string {
	return s.addr
}

func (s *server) Run() {
	hs := &http.Server{
		Addr:    s.addr,
		Handler: s,
	}

	s.log.Fatal(hs.ListenAndServe())
}

func (s *server) RequestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.count
}

func (s *server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = 0
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.log.Println("received request")
	s.incrementCount()

	ctx := r.Context()

	var depth int64
	if s.nextServerPort > 0 {
		serverDepth, err := client.GetDepth(ctx, s.nextServerPort)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		depth = serverDepth + 1
	} else if s.serverErrResp {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		if s.reqTimeout {
			<-ctx.Done()
		}

		depth = 1
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, depth)
}

func (s *server) incrementCount() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = s.count + 1
}
