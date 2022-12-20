package server

import (
	"errors"
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
	statusCode     int
	// let the request timeout to simulate a server under extreme load
	reqTimeout bool
	mu         sync.Mutex
	count      int
}

func NewIntermediateServer(port, nextServerPort int) Server {
	return newServer(port, nextServerPort, http.StatusOK, false)
}

func NewTerminalServer(port, statusCode int, reqTimeout bool) Server {
	return newServer(port, 0, statusCode, reqTimeout)
}

func newServer(port, nextServerPort, statusCode int, reqTimeout bool) *server {
	addr := fmt.Sprintf(":%d", port)
	logger := log.New(os.Stderr, fmt.Sprintf("[%s] ", addr), log.Ltime)

	return &server{addr, logger, nextServerPort, statusCode, reqTimeout, sync.Mutex{}, 0}
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
	requestChain := r.URL.Query().Get("requestChain")

	rc := s.incrementCount()
	s.log.Printf("received request %s (request #%d)\n", requestChain, rc)

	ctx := r.Context()

	var depth int64
	if s.nextServerPort > 0 {
		serverDepth, err := client.GetDepth(ctx, s.nextServerPort, requestChain)
		if err != nil {
			var respCodeErr client.ResponseCodeError
			// propagate 4xx responses
			if errors.As(err, &respCodeErr) && respCodeErr.StatusRange == 4 {
				w.WriteHeader(respCodeErr.ResponseCode)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		depth = serverDepth + 1
	} else {
		if s.reqTimeout {
			<-ctx.Done()
		}

		depth = 1
	}

	w.WriteHeader(s.statusCode)
	fmt.Fprint(w, depth)
}

func (s *server) incrementCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++
	return s.count
}
