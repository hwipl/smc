# SMC Socket Programming

## C

See [csmc.c](csmc.c) for a C program that runs a client or server with an
AF\_SMC socket.

SMC protocol definitions:

```c
// #define AF_SMC 43
#define SMCPROTO_SMC 0
#define SMCPROTO_SMC6 1
```

SMC socket address (AF\_INET and AF\_INET6) creation code:

```c
/* create an ipv4 socket address with address and port */
int create_sockaddr4(char *address, int port, struct sockaddr_in *sockaddr) {
        /* set ipv4 defaults */
        sockaddr->sin_family = AF_INET;
        sockaddr->sin_addr.s_addr = INADDR_ANY;
        sockaddr->sin_port = htons(PORT);

        /* parse given port if specified */
        if (port > 0) {
                sockaddr->sin_port = htons(port);
        }

        /* parse given address if specified */
        if (address && inet_pton(AF_INET, address, &sockaddr->sin_addr) != 1) {
                return -1;
        }
        return sizeof(struct sockaddr_in);
}

/* create an ipv6 socket address with address and port */
int create_sockaddr6(char *address, int port, struct sockaddr_in6 *sockaddr) {
        /* set ipv6 defaults */
        sockaddr->sin6_family = AF_INET6;
        sockaddr->sin6_addr = in6addr_any;
        sockaddr->sin6_port = htons(PORT);

        /* parse given port if specified */
        if (port > 0) {
                sockaddr->sin6_port = htons(port);
        }

        /* parse given address if specified */
        if (address &&
            inet_pton(AF_INET6, address, &sockaddr->sin6_addr) != 1) {
                return -1;
        }
        return sizeof(struct sockaddr_in6);
}

/* create an ipv4 or ipv6 sock address with address and port */
int create_sockaddr(char *address, int port, struct sockaddr_in6 *sockaddr) {
        int sockaddr_len;

        /* check if it's an ipv4 address */
        sockaddr_len = create_sockaddr4(address, port,
                                        (struct sockaddr_in *) sockaddr);
        if (sockaddr_len > 0) {
                return sockaddr_len;
        }

        /* check if it's an ipv6 address */
        sockaddr_len = create_sockaddr6(address, port, sockaddr);
        if (sockaddr_len > 0) {
                return sockaddr_len;
        }

        /* parsing error */
        return -1;
}
```

SMC specific server socket code:

```c
struct sockaddr_in6 server_addr;
struct sockaddr_in6 client_addr;
int server_sock, client_sock;
int server_addr_len;
int client_addr_len;

/* create socket address from address and port */
server_addr_len = create_sockaddr(address, port, &server_addr);
if (server_addr_len < 0) {
        printf("Error parsing server address\n");
        return -1;
}

/* create socket */
if (server_addr.sin6_family == AF_INET6) {
        server_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC6);
} else {
        server_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);
}
if (server_sock == -1) {
        printf("Error creating socket\n");
        return -1;
}

/* bind listening address and port */
if (bind(server_sock, (struct sockaddr *) &server_addr,
         server_addr_len)) {
        printf("Error binding socket\n");
        return -1;
}

/* wait for a new connection */
if (listen(server_sock, 1)) {
        printf("Error listening on socket\n");
        return -1;
}

/* accept new connection */
client_addr_len = server_addr_len;
client_sock = accept(server_sock, (struct sockaddr *) &client_addr,
                     &client_addr_len);
if (client_sock == -1) {
        printf("Error accepting connection\n");
        return -1;
}
printf("New client connection\n");

/* read from client socket, write to client socket */

close(client_sock);
close(server_sock);
```

SMC specific client socket code:

```c
struct sockaddr_in6 server_addr;
int server_addr_len;
int client_sock;

/* create socket address from address and port */
server_addr_len = create_sockaddr(address, port, &server_addr);
if (server_addr_len < 0) {
        printf("Error parsing server address\n");
        return -1;
}

/* create socket */
if (server_addr.sin6_family == AF_INET6) {
        client_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC6);
} else {
        client_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);
}
if (client_sock == -1) {
        printf("Error creating socket\n");
        return -1;
}

/* connect to server */
if (connect(client_sock, (struct sockaddr *) &server_addr,
            server_addr_len)) {
        printf("Error connecting to server\n");
        return -1;
}
printf("Connected to server\n");

/* write to socket, read from socket  */

close(client_sock);
```

Compiling the code:

```console
$ gcc csmc.c -o csmc
```

Running the code:

```console
$ ./csmc -s
New client connection
Read 12 bytes from client: Hello, world
Sent 12 bytes to client: Hello, world
```

```console
$ ./csmc -c -a 127.0.0.1
Connected to server
Sent 12 bytes to server: Hello, world
Read 12 bytes from server: Hello, world
```

## Go

See [gosmc.go](gosmc.go) for a Go program that runs a client or server with an
AF\_SMC socket.

SMC protocol definitions:

```go
const (
        SMCProtoIPv4 = 0
        SMCProtoIPv6 = 1
)
```

The Go package `net` provides high level functions `Listen()` and `Dial()` for
creating a server and connecting to a server. These functions do not support
SMC. As a workaround, low level socket functions in the package
`golang.org/x/sys/unix` can be used.

Low level SMC versions of `Listen()` and `Dial()`:

```go
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
```

Server socket code:

```go
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
```

Client socket code:

```go
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
```

Running the code:

```console
$ ./gosmc -s
New client connection
Echoed 13 bytes to client
```

```console
$ ./gosmc -c
Connected to server
Sent 13 bytes to server: Hello, world
Read 13 bytes from server: Hello, world
```

## Python

See [pysmc.py](pysmc.py) for a Python script that runs a client or server with
an AF\_SMC socket.

SMC protocol definitions:

```python
AF_SMC = 43
SMCPROTO_SMC = 0
SMCPROTO_SMC6 = 1
```

The calls to the socket module's `bind()`, `accept()`, and `connect()` raise
the error `OSError: getsockaddrarg: bad family` because the internal
`getsockaddrarg()` function in Python's socket module cannot construct the
socket address for an unknown address family, i.e., AF\_SMC. As a workaround,
the functions in Libc can be used in Python.

Libc version of `bind()`, `accept()`, and `connect()` with the socket address
structure used by these functions:

```python
# libc
LIBC_PATH = ctypes.util.find_library("c")
LIBC = ctypes.CDLL(LIBC_PATH)


class SockaddrIn(ctypes.Structure):
    """
    Socket Address for IPv4 struct for ctypes
    """
    _fields_ = [("sin_family", ctypes.c_ushort),
                ("sin_port", ctypes.c_uint16),
                ("sin_addr", ctypes.c_uint32),
                ("sin_zero", ctypes.c_ubyte * 8)]

    def __init__(self, address=0, port=0):
        super(SockaddrIn, self).__init__()
        self.sin_family = ctypes.c_ushort(socket.AF_INET)
        self.sin_port = ctypes.c_uint16(socket.htons(int(port)))
        self.sin_addr = ctypes.c_uint32(socket.htonl(int(
            ipaddress.IPv4Address(address))))

    def __len__(self):
        return ctypes.sizeof(self)

    def __repr__(self):
        addr = ipaddress.IPv4Address(socket.ntohl(self.sin_addr))
        port = socket.ntohs(self.sin_port)
        return f"{addr}:{port}"


class SockaddrIn6(ctypes.Structure):
    """
    Socket Address for IPv6 struct for ctypes
    """
    _fields_ = [("sin6_family", ctypes.c_ushort),
                ("sin6_port", ctypes.c_uint16),
                ("sin6_flowinfo", ctypes.c_uint32),
                ("sin6_addr", ctypes.c_ubyte * 16),
                ("sin6_scope_id", ctypes.c_uint32)]

    def __init__(self, address=0, port=0):
        super(SockaddrIn6, self).__init__()
        self.sin6_family = ctypes.c_ushort(socket.AF_INET6)
        self.sin6_port = ctypes.c_uint16(socket.htons(int(port)))
        self.sin6_addr = (ctypes.c_ubyte * 16)(
            *ipaddress.IPv6Address(address).packed)

    def __len__(self):
        return ctypes.sizeof(self)

    def __repr__(self):
        addr = ipaddress.IPv6Address(bytes(self.sin6_addr))
        port = socket.ntohs(self.sin6_port)
        return f"{addr}:{port}"


def create_sockaddr(address, port):
    """
    create a sockaddr
    """

    try:
        sockaddr = SockaddrIn(address, port)
    except ipaddress.AddressValueError:
        sockaddr = None

    if not sockaddr:
        try:
            sockaddr = SockaddrIn6(address, port)
        except ipaddress.AddressValueError:
            sockaddr = None

    return sockaddr


def bind(sock, sockaddr):
    """
    bind() using libc with ctypes
    """

    # bind address and port
    LIBC.bind(sock.fileno(), ctypes.byref(sockaddr), len(sockaddr))


def accept(sock, sockaddr):
    """
    accept() using libc with ctypes
    """

    # accept connection
    sockaddr_len = ctypes.c_int(len(sockaddr))
    client_sock = LIBC.accept(sock.fileno(), ctypes.byref(sockaddr),
                              ctypes.byref(sockaddr_len))
    return socket.socket(fileno=client_sock), sockaddr


def connect(sock, sockaddr):
    """
    connect() using libc with ctypes
    """

    # connect to server
    LIBC.connect(sock.fileno(), ctypes.byref(sockaddr), len(sockaddr))
```

Server socket code:

```python
# create sockaddr from host and port
server_addr = create_sockaddr(host, port)

# start server
print("Starting SMC server")
if isinstance(server_addr, SockaddrIn6):
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC6)
    client_addr = SockaddrIn6()
else:
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC)
    client_addr = SockaddrIn()
with sock:
    # sock.bind() does not work with SMC, use libc version
    bind(sock, server_addr)
    sock.listen(1)
    # sock.accept() does not work with SMC, use libc version
    conn, addr = accept(sock, client_addr)
    with conn:
        print("Connected:", addr)
        while True:
            data = conn.recv(1024)
            if not data:
                break
            conn.sendall(data)
```

Client socket code:

```python
# create server sockaddr from host and port
server_addr = create_sockaddr(host, port)

# start client
print("Starting SMC client")
if isinstance(server_addr, SockaddrIn6):
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC6)
else:
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC)
with sock:
    # sock.connect() does not work with SMC, use libc version
    connect(sock, server_addr)
    sock.sendall(b"Hello, world")
    data = sock.recv(1024)
    print('Received:', repr(data))
```

Running the code:

```console
$ ./pysmc.py -s
Starting SMC server
Connected: 127.0.0.1:47090
```

```console
$ ./pysmc.py -c -a 127.0.0.1
Starting SMC client
Received: b'Hello, world'
```
