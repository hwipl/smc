#!/usr/bin/env python3

"""
SMC server and client socket program for testing.
Based on echo client and server examples in official python documentation
(https://docs.python.org/3/library/socket.html#example)
"""

import ctypes.util
import ctypes

import ipaddress
import argparse
import socket
import sys

# SMC address family and protocol definitions, because currently there are no
# such definitions in the socket module
AF_SMC = 43
SMCPROTO_SMC = 0
SMCPROTO_SMC6 = 1

# client/server definitions
PORT = 50000

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


def check():
    """
    Check if SMC sockets work
    """

    # try to create a SMC socket, read fileno, bind an address to it, and
    # close it again
    print("Checking SMC socket")
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC)
    sock.fileno()
    # sock.bind() does not work with SMC, use libc version
    bind(sock, "0.0.0.0", PORT)
    sock.close()


def server(host, port):
    """
    Run SMC server,
    listen on host address and port
    """

    # set host address and port
    if not host:
        host = "0.0.0.0"
    if not port:
        port = PORT

    # start server
    print("Starting SMC server")
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


def client(host, port):
    """
    Run SMC client,
    connect to host address and port
    """

    # set host address and port
    if not host:
        host = "127.0.0.1"
    if not port:
        port = PORT

    # start client
    print("Starting SMC client")
    with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
        # sock.connect() does not work with SMC, use libc version
        connect(sock, host, port)
        sock.sendall(b"Hello, world")
        data = sock.recv(1024)
        print('Received:', repr(data))


def main():
    """
    Main
    """

    # parse command line arguments
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true",
                        help="check SMC socket")
    parser.add_argument("-s", "--server", action="store_true",
                        help="run server")
    parser.add_argument("-c", "--client", action="store_true",
                        help="run client")
    parser.add_argument("-a", "--address", help="listen/connect address")
    parser.add_argument("-p", "--port", help="listen/connect port")
    args = parser.parse_args()

    # run other commands based on command line arguments
    if args.check:
        check()
        return

    if args.client:
        client(args.address, args.port)
        return

    if args.server:
        server(args.address, args.port)
        return


if __name__ == "__main__":
    sys.exit(main())
