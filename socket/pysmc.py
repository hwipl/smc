#!/usr/bin/env python3

"""
SMC server and client socket program for testing.
Based on echo client and server examples in official python documentation
(https://docs.python.org/3/library/socket.html#example)
"""

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


def check():
    """
    Check if SMC sockets work
    """

    # try to create a SMC socket, read fileno, bind an address to it, and
    # close it again
    print("Checking SMC socket")
    sock = socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC)
    sock.fileno()
    sock.bind(("", PORT))
    sock.close()


def server(host, port):
    """
    Run SMC server,
    listen on host address and port
    """

    # set host address and port
    if not host:
        host = ""
    if not port:
        port = PORT

    # start server
    print("Starting SMC server")
    with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
        sock.bind((host, port))
        sock.listen(1)
        conn, addr = sock.caccept()
        with conn:
            print("Connected: ", addr)
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
        host = "localhost"
    if not port:
        port = PORT

    # start client
    print("Starting SMC client")
    with socket.socket(AF_SMC, socket.SOCK_STREAM, SMCPROTO_SMC) as sock:
        sock.connect((host, port))
        sock.sendall(b'Hello, world')
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
