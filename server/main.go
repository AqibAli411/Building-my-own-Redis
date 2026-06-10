package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server running on :1234")
	for {
		_, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
		}
	}
}
