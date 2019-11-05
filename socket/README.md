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

SMC specific server socket code:

```c
struct sockaddr_in server_addr;
int server_sock, client_sock;

server_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);

server_addr.sin_family = AF_INET;
server_addr.sin_addr.s_addr = INADDR_ANY;
server_addr.sin_port = htons(PORT);

bind(server_sock, (struct sockaddr *) &server_addr, sizeof(server_addr));
listen(server_sock, 1);
client_sock = accept(server_sock, NULL, NULL);

/* read from socket, write to socket */

close(client_sock);
close(server_sock);
```

SMC specific client socket code:

```c
struct sockaddr_in server_addr;
int client_sock;

client_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);

server_addr.sin_family = AF_INET;
inet_pton(AF_INET, address, &server_addr.sin_addr);
server_addr.sin_port = htons(PORT);

connect(client_sock, (struct sockaddr *) &server_addr, sizeof(server_addr));

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
