package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hashtagchris/nested-retries/client"
	"github.com/hashtagchris/nested-retries/server"
)

const startingPort = 8001
const numberOfServers = 5

func main() {
	port := startingPort

	for i := 0; i < numberOfServers-1; i = i + 1 {
		server := server.NewIntermediateServer(port, port+1)
		go server.Run()

		port = port + 1
	}

	server := server.NewTerminalServer(port, false)
	go server.Run()

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
			continue
		}
		fmt.Printf("Request depth: %d\n", depth)
	}
}
