/*
 * SMC server and client socket program for testing.
 * Runs a simple echo server and client.
 */

#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <stdio.h>

/* SMC defines */
// #define AF_SMC 43
#define SMCPROTO_SMC 0
#define SMCPROTO_SMC6 1

#define PORT 50000

int create_sockaddr4(char *address, int port, struct sockaddr_in *sockaddr) {
	// set ipv4 defaults
	sockaddr->sin_family = AF_INET;
	sockaddr->sin_addr.s_addr = INADDR_ANY;
	sockaddr->sin_port = htons(PORT);

	/* parse given port if specified */
	if (port > 0) {
		sockaddr->sin_port = htons(port);
	}

	// parse given address if specified
	if (address && inet_pton(AF_INET, address, &sockaddr->sin_addr) != 1) {
		return -1;
	}
	return 0;
}

int create_sockaddr6(char *address, int port, struct sockaddr_in6 *sockaddr) {
	// set ipv6 defaults
	sockaddr->sin6_family = AF_INET6;
	sockaddr->sin6_addr = in6addr_any;
	sockaddr->sin6_port = htons(PORT);

	/* parse given port if specified */
	if (port > 0) {
		sockaddr->sin6_port = htons(port);
	}

	// parse given address if specified
	if (address &&
	    inet_pton(AF_INET6, address, &sockaddr->sin6_addr) != 1) {
		return -1;
	}
	return 0;
}

/* run server */
int run_server(char *address, int port) {
	struct sockaddr_in6 server_addr6;
	struct sockaddr_in6 client_addr6;
	struct sockaddr_in server_addr4;
	struct sockaddr_in client_addr4;
	int server_sock, client_sock;
	int client_addrlen;
	char buffer[1024];
	int send_count;
	int recv_count;
	int ipv4 = 1;
	int rc;

	/* create socket address from address and port */
	if (create_sockaddr4(address, port, &server_addr4)) {
		ipv4 = 0;
		if (create_sockaddr6(address, port, &server_addr6)) {
			printf("Error parsing server address\n");
			return -1;
		}
	}

	/* create socket */
	if (ipv4) {
		server_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);
	} else {
		server_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC6);
	}
	if (server_sock == -1) {
		printf("Error creating socket\n");
		return -1;
	}

	/* bind listening address and port */
	if (ipv4) {
		rc = bind(server_sock, (struct sockaddr *) &server_addr4,
			  sizeof(server_addr4));
	} else {
		rc = bind(server_sock, (struct sockaddr *) &server_addr6,
			  sizeof(server_addr6));
	}
	if (rc) {
		printf("Error binding socket\n");
		return -1;
	}

	/* wait for a new connection */
	if (listen(server_sock, 1)) {
		printf("Error listening on socket\n");
		return -1;
	}

	/* accept new connection */
	if (ipv4) {
		client_addrlen = sizeof(client_addr4);
		client_sock = accept(server_sock,
				     (struct sockaddr *) &client_addr4,
				     &client_addrlen);
	} else {
		client_addrlen = sizeof(client_addr6);
		client_sock = accept(server_sock,
				     (struct sockaddr *) &client_addr6,
				     &client_addrlen);
	}
	if (client_sock == -1) {
		printf("Error accepting connection\n");
		return -1;
	}
	printf("New client connection\n");

	/* read from client */
	recv_count = -1;
	while (recv_count) {
		/* clear buffer */
		memset(buffer, 0, sizeof(buffer));

		/* receive as much as possible from client */
		recv_count = read(client_sock, buffer, sizeof(buffer));
		if (recv_count < 0) {
			printf("Error reading from connection\n");
			return -1;
		}
		if (!recv_count) {
			/* EOF reached, stop */
			return 0;
		}
		printf("Read %d bytes from client: %s\n", recv_count, buffer);

		/* send everything back to client */
		send_count = 0;
		while (send_count < recv_count) {
			send_count += send(client_sock, buffer + send_count,
					   recv_count - send_count, 0);
		}
		printf("Sent %d bytes to client: %s\n", send_count, buffer);
	}

	close(client_sock);
	close(server_sock);

	return 0;
}

/* run client */
int run_client(char *address, int port) {
	struct sockaddr_in6 server_addr6;
	struct sockaddr_in server_addr4;
	const char *hello = "Hello, world";
	char buffer[1024];
	int client_sock;
	int count;
	int ipv4;
	int rc;

	/* create socket address from address and port */
	if (create_sockaddr4(address, port, &server_addr4)) {
		ipv4 = 0;
		if (create_sockaddr6(address, port, &server_addr6)) {
			printf("Error parsing server address\n");
			return -1;
		}
	}
	/* create socket */
	if (ipv4) {
		client_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC);
	} else {
		client_sock = socket(AF_SMC, SOCK_STREAM, SMCPROTO_SMC6);
	}
	if (client_sock == -1) {
		printf("Error creating socket\n");
		return -1;
	}

	/* connect to server */
	if (ipv4) {
		rc = connect(client_sock, (struct sockaddr *) &server_addr4,
			     sizeof(server_addr4));
	} else {
		rc = connect(client_sock, (struct sockaddr *) &server_addr6,
			     sizeof(server_addr6));
	}
	if (rc) {
		printf("Error connecting to server\n");
		return -1;
	}
	printf("Connected to server\n");

	/* send message to server */
	count = 0;
	while (count < strlen(hello)) {
		count += send(client_sock, hello + count,
			      strlen(hello) - count, 0);
	}
	printf("Sent %d bytes to server: %s\n", count, hello);

	/* read from server */
	count = 0;
	while (count < strlen(hello)) {
		/* clear buffer */
		memset(buffer, 0, sizeof(buffer));

		count += read(client_sock, buffer, sizeof(buffer));
		if (count < 0) {
			printf("Error reading from connection\n");
			return -1;
		}
		printf("Read %d bytes from server: %s\n", count, buffer);
	}

	close(client_sock);

	return 0;
}

/* print usage */
void print_usage() {
	printf("Usage:\n"
	       "    csmc -c -a <address> [-p <port>]\n"
	       "        run client and connect to address and port\n"
	       "    csmc -s [-a <address>] [-p <port>]\n"
	       "        run server and listen on address and port\n"
	       );
}

/* main function */
int main(int argc, char** argv) {
	int client = 0, server = 0, port = 0;
	char *address = NULL;
	int c;

	/* parse command line arguments */
	while ((c = getopt (argc, argv, "a:cp:s")) != -1) {
		switch (c) {
		case 'c':
			/* client */
			client = 1;
			break;
		case 's':
			/* server */
			server = 1;
			break;
		case 'a':
			/* listening/connecting address */
			address = optarg;
			break;
		case 'p':
			/* listening/connecting port */
			port = atoi(optarg);
			break;
		default:
			/* wrong command line argument */
			print_usage();
			return 0;
		}
	}

	if (server) {
		/* server mode */
		return run_server(address, port);
	}

	if (client) {
		/* client mode */
		if (address) {
			return run_client(address, port);
		}
	}

	/* wrong/missing command line arguments */
	print_usage();
	return 0;
}
