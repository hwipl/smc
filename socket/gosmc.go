package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/sys/unix"
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
	SMCProtoIPv4 = 0
	SMCProtoIPv6 = 1
)

// run as a server
func runServer(address string, port int) {
	// create socket
	fd, err := unix.Socket(unix.AF_SMC, unix.SOCK_STREAM, SMCProtoIPv4)
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
			fmt.Printf("New client connection\n")
			written, err := io.Copy(c, c)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Echoed %d bytes to client\n", written)
			c.Close()
		}(conn)
	}
}

// run as a client
func runClient(address string, port int) {
	// create socket
	fd, err := unix.Socket(unix.AF_SMC, unix.SOCK_STREAM, SMCProtoIPv4)
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
	fmt.Printf("Connected to server\n")

	// create a connection from connected socket
	file := os.NewFile(uintptr(fd), "")
	conn, err := net.FileConn(file)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	defer unix.Close(fd)

	// sent text, read reply an
	text := "Hello, world\n"
	fmt.Fprintf(conn, text)
	fmt.Printf("Sent %d bytes to server: %s", len(text), text)
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Read %d bytes from server: %s", len(reply), reply)
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
		runServer(address, port)
		return
	}

	// run client
	if client {
		runClient(address, port)
		return
	}
}
