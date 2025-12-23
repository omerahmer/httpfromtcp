package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("err", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		request.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})

		fmt.Printf("Body:\n")
		fmt.Printf("%s\n", request.Body)
	}
}
