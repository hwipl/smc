# SMC Socket Programming

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
