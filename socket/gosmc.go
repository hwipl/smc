package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"net"
	"os"
)

// variable definitions
var (
	server  bool   = false
	client  bool   = false
	address string = "127.0.0.1"
	port    int    = 50000
)

// SMC definitions
const (
	SMCPROTO_SMC  = 0
	SMCPROTO_SMC6 = 1
)

// run as a server
func run_server(address string, port int) {
	// create socket
	fd, err := unix.Socket(unix.AF_SMC, unix.SOCK_STREAM, SMCPROTO_SMC)
	if err != nil {
		log.Fatal(err)
	}

	// construct socket address
	sockaddr := &unix.SockaddrInet4{}
	sockaddr.Port = port
	ip := net.ParseIP(address).To4()
	if ip == nil {
		log.Fatal("Error parsing IP")
	}
	copy(sockaddr.Addr[:], ip[0:4])

	// bind socket address
	err = unix.Bind(fd, sockaddr)
	if err != nil {
		log.Fatal(err)
	}

	// start listening
	err = unix.Listen(fd, 1)
	if err != nil {
		log.Fatal(err)
	}


	// create a listener from listening socket
	file := os.NewFile(uintptr(fd), "")
	l, err := net.FileListener(file)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	defer unix.Close(fd)

	// accept new connections from listener and handle them
	for {
		// accept new connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// handle new connection: send data back to client
		go func(c net.Conn) {
			io.Copy(c, c)
			c.Close()
		}(conn)
	}
}

// run as a client
func run_client(address string, port int) {
	// create socket
	fd, err := unix.Socket(unix.AF_SMC, unix.SOCK_STREAM, SMCPROTO_SMC)
	if err != nil {
		log.Fatal(err)
	}

	// construct socket address
	sockaddr := &unix.SockaddrInet4{}
	sockaddr.Port = port
	ip := net.ParseIP(address).To4()
	if ip == nil {
		log.Fatal("Error parsing IP")
	}
	copy(sockaddr.Addr[:], ip[0:4])

	// connect to server
	err = unix.Connect(fd, sockaddr)
	if err != nil {
		log.Fatal(err)
	}

	// create a connection from connected socket
	file := os.NewFile(uintptr(fd), "")
	conn, err := net.FileConn(file)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	defer unix.Close(fd)

	// sent text, read reply an
	fmt.Fprintf(conn, "Hello, world\n")
	reply, err := bufio.NewReader(conn).ReadString('\n')
	fmt.Printf("%s", reply)
}

func main() {
	// parse command line arguments
	flag.BoolVar(&server, "s", server, "run server")
	flag.BoolVar(&client, "c", client, "run client")
	flag.StringVar(&address, "a", address, "server/client address")
	flag.IntVar(&port, "p", port, "server/client port")

	flag.Parse()

	// run server
	if server {
		run_server(address, port)
		return
	}

	// run client
	if client {
		run_client(address, port)
		return
	}
}
