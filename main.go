package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hashtagchris/nested-retries/client"
	"github.com/hashtagchris/nested-retries/server"
)

const startingPort = 8001
const numberOfServers = 3
const lastServerReturnsErr = true

var servers []server.Server

func main() {
	port := startingPort

	for i := 0; i < numberOfServers-1; i = i + 1 {
		servers = append(servers, server.NewIntermediateServer(port, port+1))
		port = port + 1
	}

	servers = append(servers, server.NewTerminalServer(port, lastServerReturnsErr))

	for _, server := range servers {
		go server.Run()
	}

	makeRequests()
}

func makeRequests() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println()
		fmt.Println("Hit enter to make a http request")
		if !scanner.Scan() {
			return
		}
		depth, err := client.GetDepth(startingPort)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			fmt.Printf("Success! Request depth: %d\n", depth)
		}

		fmt.Println()
		for _, server := range servers {
			server.LogRequestCount()
			server.Reset()
		}
	}
}
