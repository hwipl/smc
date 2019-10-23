# SMC Socket Programming

## C

See `csmc.c` for a C program that runs a client or server with an AF\_SMC
socket.

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

See `pysmc.py` for a **not working** Python script that runs a client or server
with an AF\_SMC socket.

SMC protocol definitions:

```python
AF_SMC = 43
SMCPROTO_SMC = 0
SMCPROTO_SMC6 = 1
```

Server socket code:

```python
with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
    sock.bind((host, port))
    sock.listen(1)
    conn, addr = sock.accept()
    with conn:
        print("Connected: ", addr)
        while True:
            data = conn.recv(1024)
            if not data:
                break
            conn.sendall(data)
```

Client socket code:

```python
with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
    sock.connect((host, port))
    sock.sendall(b"Hello, world")
    data = sock.recv(1024)
    print('Received:', repr(data))
```

Running the code:

```console
$ ./pysmc.py --server
[...]
  File "./pysmc.py", line 52, in server
    sock.bind((host, port))
OSError: getsockaddrarg: bad family
$ ./pysmc.py --client
[...]
  File "./pysmc.py", line 79, in client
    sock.connect((host, port))
OSError: getsockaddrarg: bad family
```

The calls to `bind()` and `connect()` raise the error `OSError: getsockaddrarg:
bad family` because the internal `getsockaddrarg()` function in Python's socket
module cannot construct the socket address for an unknown address family, i.e.,
AF\_SMC.
