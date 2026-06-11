package main

import (
	"fmt"
	"net"
	"os"
	"golang.org/x/sys/unix"
)

func fb_set_nb(fb int) {
	flags, err := unix.FcntlInt(uintptr(fb), unix.F_GETFL, 0)
	if err != nil {
		panic(err)
	}
	flags |= unix.O_NONBLOCK
	_, err = unix.FcntlInt(uintptr(fb), unix.F_SETFL, flags)
	if err != nil {
		panic(err)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server running on :1234")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	cmd, err := ParseReq(buf[:n])
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	var out []byte
	DoRequest(cmd, &out)
	conn.Write(out)
}
