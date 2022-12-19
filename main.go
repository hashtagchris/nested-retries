package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashtagchris/nested-retries/client"
	"github.com/hashtagchris/nested-retries/server"
)

const startingPort = 8001
const intermediateServers = 2
const terminalServerStatusCode = http.StatusOK

// let requests to the terminal server timeout, simulating extreme load
const terminalServerTimeouts = true

var servers []server.Server
var terminalServer server.Server

func main() {
	port := startingPort

	for i := 0; i < intermediateServers; i = i + 1 {
		servers = append(servers, server.NewIntermediateServer(port, port+1))
		port = port + 1
	}

	terminalServer = server.NewTerminalServer(port, terminalServerStatusCode, terminalServerTimeouts)
	servers = append(servers, terminalServer)

	for _, server := range servers {
		go server.Run()
	}

	makeRequests(context.Background())
}

func makeRequests(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println()
		fmt.Println("Hit enter to make a http request")
		if !scanner.Scan() {
			return
		}

		startAt := time.Now()
		depth, err := client.GetDepth(ctx, startingPort)
		elapsed := time.Since(startAt)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Printf("Success! Request depth: %d\n", depth)
		}

		for _, server := range servers {
			fmt.Printf("Requests received by %s server: %d\n", server.ID(), server.RequestCount())
			server.Reset()
		}
		fmt.Printf("Elapsed sec: %d\n", int64(elapsed.Seconds()))
		fmt.Println()
	}
}
