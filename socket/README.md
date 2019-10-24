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


def bind(sock, address, port):
    """
    bind() using libc with ctypes
    """

    # bind address and port
    sockaddr = SockaddrIn(address, port)
    LIBC.bind(sock.fileno(), ctypes.byref(sockaddr), len(sockaddr))


def accept(sock):
    """
    accept() using libc with ctypes
    """

    # accept connection
    sockaddr = SockaddrIn()
    sockaddr_len = ctypes.c_int(len(sockaddr))
    client_sock = LIBC.accept(sock.fileno(), ctypes.byref(sockaddr),
                              ctypes.byref(sockaddr_len))
    return socket.socket(fileno=client_sock), sockaddr


def connect(sock, address, port):
    """
    connect() using libc with ctypes
    """

    # connect to server
    sockaddr = SockaddrIn(address, port)
    LIBC.connect(sock.fileno(), ctypes.byref(sockaddr), len(sockaddr))
```

Server socket code:

```python
with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
    # sock.bind() does not work with SMC, use libc version
    bind(sock, host, port)
    sock.listen(1)
    # sock.accept() does not work with SMC, use libc version
    conn, addr = accept(sock)
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
with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
    # sock.connect() does not work with SMC, use libc version
    connect(sock, host, port)
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
