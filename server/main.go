package main

import (
	"fmt"
	"log"
	"golang.org/x/sys/unix"
)

const (
	STATE_REQ = 0
	STATE_RES = 1
	STATE_END = 2
)

type Conn struct {
	fd        int
	state     int
	rbuf_size int
	rbuf      [4 + 4096]byte
	wbuf_sent int
	wbuf_size int
	wbuf      [4 + 4096]byte
}

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

func accept_conn(fd int, fd2Conn map[int32]*Conn) error {
	connFd, _, err := unix.Accept(fd)
	if err != nil {
		return err
	}
	fb_set_nb(connFd)
	conn := &Conn{
		fd:    connFd,
		state: STATE_REQ,
	}
	fd2Conn[int32(connFd)] = conn
	return nil
}

func main() {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)
	
	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	addr := &unix.SockaddrInet4{Port: 1234, Addr: [4]byte{0, 0, 0, 0}}
	unix.Bind(fd, addr)
	unix.Listen(fd, unix.SOMAXCONN)
	
	fmt.Println("Server running on :1234")
	fd2Conn := make(map[int32]*Conn)
	fb_set_nb(fd)
	
	var poll_args []unix.PollFd
	for {
		poll_args = []unix.PollFd{
			{Fd: int32(fd), Events: unix.POLLIN},
		}
		for _, conn := range fd2Conn {
			var events int16 = unix.POLLERR
			if conn.state == STATE_REQ {
				events |= unix.POLLIN
			} else {
				events |= unix.POLLOUT
			}
			poll_args = append(poll_args, unix.PollFd{Fd: int32(conn.fd), Events: events})
		}
		
		_, err := unix.Poll(poll_args, -1)
		if err != nil {
			panic(err)
		}
		
		for i := 1; i < len(poll_args); i++ {
			if poll_args[i].Revents == 0 {
				continue
			}
			conn := fd2Conn[poll_args[i].Fd]
			if conn == nil {
				continue
			}
			connectionIO(conn, fd2Conn)
		}
		
		if poll_args[0].Revents != 0 {
			accept_conn(fd, fd2Conn)
		}
	}
}

func connectionIO(conn *Conn, fd2Conn map[int32]*Conn) {
	if conn.state == STATE_REQ {
		state_req(conn)
	} else if conn.state == STATE_RES {
		state_res(conn)
	}
	if conn.state == STATE_END {
		unix.Close(conn.fd)
		delete(fd2Conn, int32(conn.fd))
	}
}

func state_req(conn *Conn) {
	n, err := unix.Read(conn.fd, conn.rbuf[conn.rbuf_size:])
	if n < 0 && err == unix.EAGAIN {
		return
	}
	if n <= 0 {
		conn.state = STATE_END
		return
	}
	conn.rbuf_size += n
	for try_one_request(conn) {}
}

func try_one_request(conn *Conn) bool {
	if conn.rbuf_size < 4 {
		return false
	}
	// Simplified parsing for now: echo back or parse basic string
	conn.state = STATE_RES
	copy(conn.wbuf[:], conn.rbuf[:conn.rbuf_size])
	conn.wbuf_size = conn.rbuf_size
	conn.rbuf_size = 0
	return false
}

func state_res(conn *Conn) {
	n, err := unix.Write(conn.fd, conn.wbuf[conn.wbuf_sent:conn.wbuf_size])
	if n < 0 && err == unix.EAGAIN {
		return
	}
	if n <= 0 {
		conn.state = STATE_END
		return
	}
	conn.wbuf_sent += n
	if conn.wbuf_sent == conn.wbuf_size {
		conn.state = STATE_REQ
		conn.wbuf_size = 0
		conn.wbuf_sent = 0
	}
}
