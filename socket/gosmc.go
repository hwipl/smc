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

// parse ip address and return IPv4 and IPv6 address
func parseIP(address string) (net.IP, net.IP) {
	ip := net.ParseIP(address)
	if ip == nil {
		return nil, nil
	}
	return ip.To4(), ip.To16()
}

// construct socket address
func createSockaddr(address string, port int) (typ string, s unix.Sockaddr) {
	ipv4, ipv6 := parseIP(address)
	if ipv4 != nil {
		sockaddr4 := &unix.SockaddrInet4{}
		sockaddr4.Port = port
		copy(sockaddr4.Addr[:], ipv4[:net.IPv4len])
		return "ipv4", sockaddr4
	}
	if ipv6 != nil {
		sockaddr6 := &unix.SockaddrInet6{}
		sockaddr6.Port = port
		copy(sockaddr6.Addr[:], ipv6[:net.IPv6len])
		return "ipv6", sockaddr6
	}

	return "err", nil
}

// SMC version of Listen()
func smcListen(address string, port int) (net.Listener, error) {
	var l net.Listener
	var err error
	var fd int

	// construct socket address from address and port
	typ, sockaddr := createSockaddr(address, port)
	if typ == "err" {
		return l, fmt.Errorf("Error parsing IP")
	}

	// create socket
	if typ == "ipv4" {
		fd, err = unix.Socket(unix.AF_SMC, unix.SOCK_STREAM,
			SMCProtoIPv4)
	} else {
		fd, err = unix.Socket(unix.AF_SMC, unix.SOCK_STREAM,
			SMCProtoIPv6)
	}
	if err != nil {
		return l, err
	}
	defer unix.Close(fd)

	// bind socket address
	err = unix.Bind(fd, sockaddr)
	if err != nil {
		return l, err
	}

	// start listening
	err = unix.Listen(fd, 1)
	if err != nil {
		return l, err
	}

	// create a listener from listening socket
	file := os.NewFile(uintptr(fd), "")
	l, err = net.FileListener(file)
	return l, err
}

// SMC version of Dial()
func smcDial(address string, port int) (net.Conn, error) {
	var conn net.Conn
	var err error
	var fd int

	// construct socket address from address and port
	typ, sockaddr := createSockaddr(address, port)
	if typ == "err" {
		return conn, fmt.Errorf("Error parsing IP")
	}

	// create socket
	if typ == "ipv4" {
		fd, err = unix.Socket(unix.AF_SMC, unix.SOCK_STREAM,
			SMCProtoIPv4)
	} else {
		fd, err = unix.Socket(unix.AF_SMC, unix.SOCK_STREAM,
			SMCProtoIPv6)
	}

	if err != nil {
		return conn, err
	}
	defer unix.Close(fd)

	// connect to server
	err = unix.Connect(fd, sockaddr)
	if err != nil {
		return conn, err
	}

	// create a connection from connected socket
	file := os.NewFile(uintptr(fd), "")
	conn, err = net.FileConn(file)
	return conn, err
}

// run as a server
func runServer(address string, port int) {
	// listen for smc connections
	l, err := smcListen(address, port)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

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
	// connect via smc
	conn, err := smcDial(address, port)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Printf("Connected to server\n")

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
